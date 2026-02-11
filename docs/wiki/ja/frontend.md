# フロントエンド

## 実行時シェル構成

frontendは単一Reactアプリで、ルートprefixによりアプリ責務を切り替えます。
共通シェル（`Layout`）がナビゲーションと操作導線を管理します。

## ルート全体像

ルート定義は `frontend/src/routes.tsx` に集約されています。

```text
/                          ホームダッシュボード
/attendance, /leaves...    勤怠ドメイン
/expenses/*                経費ドメイン
/hr/*                      人事ドメイン
/wiki/*                    社内Wiki
/login                     ログイン
```

## アプリ切替モデル

`frontend/src/config/apps.ts` のアプリ定義を基準に `getActiveApp(pathname)` で現在アプリを判定します。

```text
pathname -> getActiveApp -> Layoutナビ生成 -> AppSwitcher表示
```

## Layoutの責務

`frontend/src/components/layout/Layout.tsx` は以下を担当します。

- desktop/mobileナビゲーション
- アプリ別サイドメニュー切替
- 言語切替（`ja` / `en`）
- テーマ切替（`system` -> `light` -> `dark`）
- ログアウト導線

## 社内Wikiの描画方式

`frontend/src/pages/wiki/WikiPage.tsx` の仕様:

- 英語: `docs/wiki/*.md` を読み込み
- 日本語: `docs/wiki/ja/*.md` を読み込み
- `i18n.language` で表示言語を切替
- 見出し・段落・箇条書き・コードブロックを描画

## APIクライアント構成

`frontend/src/api/client.ts` は2層構成です。

1. `openapi-fetch` クライアント（interceptor付き）
2. ドメイン別メソッド群（`api.attendance`, `api.expenses`, `api.hr` など）

### トークン更新シーケンス

```text
認証付きAPI呼び出し
  -> 401応答
  -> /auth/refresh 実行
  -> 成功: token更新
  -> 元リクエストを1回再試行
  -> 失敗: logoutして /login へ遷移
```

## 状態管理

- 認証状態: `frontend/src/stores/authStore.ts`（Zustand + persist）
- テーマ状態: `frontend/src/stores/themeStore.ts`
- サーバ状態: TanStack Query
- 画面遷移状態: TanStack Router

## i18n戦略

- 初期化: `frontend/src/i18n/index.ts`
- 辞書: `frontend/src/i18n/locales/ja.json`, `en.json`
- デフォルト言語: 日本語（`ja`）
- Wiki本文は言語別Markdownソースで管理

## ビルドと配信

`frontend/vite.config.ts` の主要設定:

- dev server: `3000`
- `/api` を backend `http://localhost:8080` へproxy
- manualChunksで vendor/router/query/ui/i18n を分割

## 画面追加プレイブック

1. `frontend/src/pages` 配下にページ作成
2. `routes.tsx` にルート追加
3. 必要なら `Layout` のナビ項目追加
4. API client・query/mutationを追加
5. テスト追加（コンポーネント/ルート挙動）
6. 仕様変更があればWiki更新

## アプリ追加プレイブック

1. `frontend/src/config/apps.ts` にアプリ定義追加
2. `routes.tsx` にルート群追加
3. `Layout.tsx` でアプリ別ナビを追加
4. ページ群・API呼び出しを実装
5. AppSwitcher/レイアウトのテスト追加

## 品質チェック項目

- ルートprefixと責務境界が明確
- 401時のrefresh/retryが破綻しない
- 言語切替でラベルとWiki本文が同期
- モバイル/デスクトップ双方で導線が成立
- buildが型エラーなく通過

## よく使うコマンド

```sh
cd frontend
pnpm dev
pnpm test
pnpm build
pnpm test:e2e
```


