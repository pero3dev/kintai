# テスト

## テスト戦略

変更範囲とリスクに応じて、層を使い分けます。

```text
          +-----------------------------+
          | E2E（主要ユーザフロー）      |
          +-----------------------------+
          | 統合/分岐テスト              |
          +-----------------------------+
          | 単体テスト（業務ロジック）     |
          +-----------------------------+
```

## スコープ選定ルール

- 軽微なUI文言変更: frontendの対象コンポーネントテスト
- 業務ルール変更: backendのservice/handler/repositoryテスト
- 認証/権限/ルート変更: router/middleware分岐テスト
- 複数画面に跨る変更: Playwright E2Eを追加・更新

## Backendテスト層

- serviceテスト: 業務分岐、ユースケース検証
- handlerテスト: 入力バインド、HTTPステータス、応答形式
- repositoryテスト: クエリ挙動、DB境界ケース
- router/middlewareテスト: 認証、ロール制限、到達性
- backend結合テスト: 実DB + middleware + handler + service + repository の経路検証

実装項目一覧:

- `docs/wiki/ja/backend-integration-test-checklist.md`
- `docs/wiki/ja/non-functional-test-checklist.md`

代表例:

- `backend/internal/router/router_branch_test.go`
- `backend/internal/middleware/middleware_branch_test.go`
- `backend/internal/service/*_test.go`

## Frontendテスト層

- componentテスト: 描画、操作、state遷移
- layout/navigationテスト: アプリ切替、ロール別表示、言語切替
- pageテスト: ルート固有の表示と挙動
- storeテスト: 認証・テーマなど永続状態

代表例:

- `frontend/src/components/layout/Layout.test.tsx`
- `frontend/src/pages/wiki/WikiPage.test.tsx`
- `frontend/src/config/apps.test.ts`

## E2E設定

Playwright設定: `frontend/playwright.config.ts`

- baseURL: `http://localhost:3000`
- webServer command: `pnpm dev`
- browser project: chromium
- retry時にtrace収集

## 負荷テスト設定（k6）

シナリオ管理: `backend/loadtest/k6`

- `scenarios/load-profile.js`: 通常負荷→ピーク→スパイク
- `scenarios/high-concurrency.js`: 多数同時ユーザー/セッション
- `scenarios/soak-endurance.js`: 長時間耐久

## CIで実行される検証

ワークフロー: `.github/workflows/ci.yml`

- backend lint
- backend test（Postgres/Redis service container付き）
- frontend lint + type check
- frontend test + coverage
- frontend Playwright E2E
- mainブランチでdocker build

定期負荷試験ワークフロー: `.github/workflows/load-test.yml`

- 実行トリガー: `schedule`（毎週）/ `workflow_dispatch`
- 実行内容: k6 `load-profile` / `high-concurrency` / `soak-endurance`
- 結果: k6 summary JSON と backendログを artifact 保存

## 実行コマンド

```sh
# frontend
cd frontend
pnpm test
pnpm test:coverage
pnpm build
pnpm test:e2e

# backend
cd backend
go test ./... -v -race -coverprofile=coverage.out -covermode=atomic

# shortcut
make backend-test
make frontend-test
make frontend-e2e
make loadtest-k6-load-profile
make loadtest-k6-high-concurrency
make loadtest-k6-soak
```

## テスト作成ガイド

- 実装詳細よりも振る舞いをテスト名に表現する
- 成功系だけでなく失敗系を必ず含める
- 時刻依存データは明示して決定的にする
- モックは境界に限定し、過剰モックを避ける

## PR前チェック

1. 変更ファイルに対応するテストをローカルで実行
2. frontend/backendのbuildを確認
3. バグ修正時は再発防止テストを追加
4. 未カバーのリスクがあればPRに明記

## Flaky調査観点

- タイムゾーン/時刻依存の前提崩れ
- テスト間での共有状態汚染
- 非同期待機不足やretry条件不備
- 外部依存（network/fs/env）のリーク
- Playwright trace/screenshotの確認

## 完了基準

- 必須CIジョブがすべてgreen
- 理由不明のskipがない
- 新機能に少なくとも1件の振る舞いテストがある
- アーキテクチャ変更時はWiki更新が完了している


