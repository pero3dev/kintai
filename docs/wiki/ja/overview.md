# 概要

## このWikiの目的

この社内Wikiは、勤怠管理システムの技術情報を一箇所に集約するためのドキュメントです。
「どこを触ればいいか分からない」状態から「安全に変更できる」状態まで、最短で到達できることを目的にしています。

## 想定読者

- 新規参画したエンジニア
- 機能追加・改修を行う開発者
- 設計妥当性を確認するレビュアー
- 障害対応や運用確認を行う担当者

## 推奨読書順

- 全体理解が目的の場合:
  1. `overview.md`
  2. `architecture.md`
  3. `backend.md`
  4. `frontend.md`
  5. `infrastructure.md`
  6. `testing.md`
  7. `backend-integration-test-checklist.md`
  8. `non-functional-test-checklist.md`
- API追加の場合: `architecture.md` -> `backend.md` -> `testing.md`
- 画面追加の場合: `architecture.md` -> `frontend.md` -> `testing.md`
- 環境トラブル対応の場合: `infrastructure.md` -> `testing.md`

## プロダクト対象範囲

同一プロダクト内で以下のアプリを切り替えて利用します。

- 勤怠アプリ（`/`, `/attendance`, `/leaves`, `/overtime`, `/corrections` など）
- 経費アプリ（`/expenses/*`）
- 人事アプリ（`/hr/*`）
- 社内Wiki（`/wiki/*`）

## システム全体像

```text
[ブラウザ]
   |
   | http://localhost:3000
   v
[Frontend (Vite/Nginx)]
   |
   | /api/v1/*
   v
[Backend (Gin)] -----> [PostgreSQL]
   |
   +---------------> [Redis]
   |
   +---------------> [Prometheus / OTel / ELK]
```

## リポジトリ構成

```text
kintai/
|- backend/                 Go API・業務ロジック・永続化層
|- frontend/                React UI・ルーティング・アプリ切替
|- docs/wiki/               英語技術ドキュメント（/wikiで利用）
|- docs/wiki/ja/            日本語技術ドキュメント（/wikiで利用）
|- infrastructure/          AWS向けTerraform構成
|- monitoring/              監視系設定（Prometheus/OTel/Logstash）
|- docker-compose.yml       ローカル統合実行環境
|- Makefile                 開発/運用コマンド集
```

## 技術スタック（現行）

- Backend: Go 1.23, Gin, GORM, PostgreSQL, Redis, JWT
- Frontend: React 19, TanStack Router, TanStack Query, Zustand, i18next
- Infrastructure: Docker Compose（ローカル）, Terraform + AWS（クラウド）
- Observability: Prometheus, OpenTelemetry Collector, ELK
- CI: GitHub Actions（`.github/workflows/ci.yml`）

## 典型的な変更フロー

1. 変更対象アプリとルート責務を特定する。
2. API契約とロール制御を確認する。
3. backend（handler/service/repository）を実装する。
4. frontend（page/component/api client）を実装する。
5. 変更範囲に応じたテストを追加・更新する。
6. build/testを通してからPRを作成する。
7. 影響範囲とロールバック方針をPRに明記する。

## 完了条件（技術観点）

- 要件どおりの動作確認が完了している
- 既存テストが回帰なく通過している
- 新規仕様に対するテストが追加されている
- 認証・権限・CORSなどの基本セキュリティが維持されている
- アーキテクチャ影響がある場合、Wiki更新が同一PRに含まれる

## ドキュメント運用ルール

- 実装事実を優先し、抽象論だけにしない
- 可能な限りファイルパスを明示する
- 複数コンポーネントに跨る内容は図を入れる
- ロール依存・環境依存の仕様は必ず明記する
