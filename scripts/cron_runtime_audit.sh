#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO_BIN="${GO_BIN:-/usr/local/go/bin/go}"
PORT="${PORT:-9031}"
LOG_PATH="${ROOT_DIR}/app.log"
TMP_DIR="$(mktemp -d)"
PATH_HELPER="${ROOT_DIR}/bin/cron_path_audit"
ASSET_HELPER="${ROOT_DIR}/bin/asset_counter_audit"

cleanup() {
  rm -rf "${TMP_DIR}"
  rm -f "${PATH_HELPER}" "${ASSET_HELPER}"
}
trap cleanup EXIT

mkdir -p "${ROOT_DIR}/bin"

cat > "${TMP_DIR}/cron_path_audit.go" <<'GO'
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	godotenv "github.com/joho/godotenv"

	crons "apotek-clean/services/crons"
	services "apotek-clean/services"
)

func latestFile(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	if len(names) == 0 {
		return ""
	}
	return filepath.Join(dir, names[len(names)-1])
}

func main() {
	envPath, err := services.ResolveEnvFilePath()
	must(err)
	must(godotenv.Load(envPath))

	projectRoot, err := services.ResolveProjectRoot()
	must(err)
	backupDir, err := services.ResolveProjectDirPath(".backup_db")
	must(err)
	before := latestFile(backupDir)
	must(crons.DBDump())
	after := latestFile(backupDir)
	fmt.Println("cwd:", mustGetwd())
	fmt.Println("project_root:", projectRoot)
	fmt.Println("backup_dir:", backupDir)
	fmt.Println("before_latest:", before)
	fmt.Println("after_latest:", after)
	if backupDir != filepath.Join(projectRoot, ".backup_db") {
		panic("backup dir tidak mengarah ke root project")
	}
	if after == "" || !strings.HasPrefix(after, backupDir+string(os.PathSeparator)) {
		panic("hasil backup tidak berada di folder backup root project")
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func mustGetwd() string {
	wd, err := os.Getwd()
	must(err)
	return wd
}
GO

cat > "${TMP_DIR}/asset_counter_audit.go" <<'GO'
package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	configs "apotek-clean/configs"
	crons "apotek-clean/services/crons"
	services "apotek-clean/services"

	godotenv "github.com/joho/godotenv"
	gorm "gorm.io/gorm"
)

type BranchAsset struct {
	BranchID string
}

type CreatedAsset struct {
	ID         string
	BranchID   string
	AssetValue int
	CreatedAt  time.Time
}

func main() {
	envPath, err := services.ResolveEnvFilePath()
	must(err)
	must(godotenv.Load(envPath))
	must(configs.SetupDB())
	db := configs.DB

	var expected []BranchAsset
	must(db.Raw(`
		SELECT branch_id
		FROM products
		GROUP BY branch_id
		ORDER BY branch_id
	`).Scan(&expected).Error)

	start := time.Now().Add(-2 * time.Second)
	must(crons.AssetCounter(db))
	end := time.Now().Add(2 * time.Second)

	var created []CreatedAsset
	must(db.Raw(`
		SELECT id, branch_id, asset_value, asset_date AS created_at
		FROM daily_assets
		WHERE asset_date >= ? AND asset_date <= ?
		ORDER BY branch_id, asset_date, id
	`, start, end).Scan(&created).Error)

	expectedBranches := make([]string, 0, len(expected))
	for _, item := range expected {
		expectedBranches = append(expectedBranches, item.BranchID)
	}
	createdBranches := make([]string, 0, len(created))
	ids := make([]string, 0, len(created))
	for _, item := range created {
		createdBranches = append(createdBranches, item.BranchID)
		ids = append(ids, item.ID)
	}
	sort.Strings(expectedBranches)
	sort.Strings(createdBranches)

	fmt.Println("expected_branch_count:", len(expectedBranches))
	fmt.Println("created_row_count:", len(created))
	fmt.Println("expected_branches:", expectedBranches)
	fmt.Println("created_branches:", createdBranches)
	if len(created) > 0 {
		fmt.Println("sample_created:", created[0].ID, created[0].BranchID, created[0].AssetValue, created[0].CreatedAt.Format(time.RFC3339))
	}

	if len(created) != len(expectedBranches) {
		cleanup(db, ids)
		fmt.Fprintln(os.Stderr, "created row count does not match expected branch count")
		os.Exit(2)
	}

	cleanup(db, ids)
	fmt.Println("cleanup_deleted_ids:", len(ids))
}

func cleanup(db *gorm.DB, ids []string) {
	if len(ids) == 0 {
		return
	}
	must(db.Exec(`DELETE FROM daily_assets WHERE id IN ?`, ids).Error)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
GO

echo "[info] build helper audit cron"
"${GO_BIN}" build -o "${PATH_HELPER}" "${TMP_DIR}/cron_path_audit.go"
"${GO_BIN}" build -o "${ASSET_HELPER}" "${TMP_DIR}/asset_counter_audit.go"

echo "[info] verifikasi backup path dari root repo"
(
  cd "${ROOT_DIR}"
  "${PATH_HELPER}"
)

echo "[info] verifikasi backup path dari folder bin"
(
  cd "${ROOT_DIR}/bin"
  ./cron_path_audit
)

echo "[info] verifikasi efek data AssetCounter dengan cleanup"
(
  cd "${ROOT_DIR}"
  "${ASSET_HELPER}"
)

echo "[info] verifikasi scheduler aktif saat startup app"
(
  cd "${ROOT_DIR}"
  PORT="${PORT}" ./scripts/restart_local.sh >/dev/null
)

for _ in $(seq 1 10); do
  if grep -q "\[SCHEDULER\] Semua job terjadwal aktif!" "${LOG_PATH}"; then
    echo "[ok] scheduler startup log terdeteksi"
    echo "[ok] cron runtime audit selesai"
    exit 0
  fi
  sleep 1
done

echo "[fail] log startup scheduler tidak ditemukan di ${LOG_PATH}" >&2
exit 1
