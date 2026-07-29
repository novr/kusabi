# Kusabi — 製品仕様

*Connect Repositories. Unify Contexts.*

Kusabi は、複数の Git リポジトリを束ね、宣言と各子の文書を**観測として**一つの文脈に差し出す CLI である。親の Git 履歴に子の実体を混ぜず、Submodule の摩擦も引き受けない。

## 解く痛み

- 親リポジトリの履歴に、子リポジトリの実体が混入する
- マルチレポ開発で、エージェントへ渡す文脈を手作業で横断収集する

## 引き受ける / 引き受けない

引き受ける:

- 子の所在・役割・束ね方の宣言（`kusabi.yaml`）
- 親履歴へ子の実体が混入しないことの防衛（`.gitignore` の宣言連動除外）
- 宣言された集合への、子単位の作用（同期・状態・一括実行）
- 宣言と各子のローカル文書からの、読み取り専用の文脈集約

引き受けない:

- 子リポジトリ内部の設計・ビルド・リリースの意味理解
- 横断変更の調停、依存解決、バージョンロックの統治
- エージェント行動方針の起草（宣言・子文書の転載のみ。不足分の補完はしない）
- 親と子を一つの Git 履歴として統合すること

## ディレクトリ構造

```plaintext
my-project-meta/
├── kusabi.yaml          # 束ねの宣言（必須）
├── AGENTS.md            # 全体方針（任意。context で観測）
├── .gitignore           # 宣言変更時に子パスを除外（init/add/remove）
└── packages/            # 子リポジトリの実体（親履歴から除外）
    ├── app-ios/
    └── app-backend/
```

## マニフェスト（kusabi.yaml）

```yaml
version: "1"
name: "my-ecosystem"
description: "Cross-platform ecosystem bound by Kusabi."

context:
  agents: "./AGENTS.md"
  paths:
    - "team-knowledge/ADR.md"
    - ".agents/skills/deploy.md"
  includes:
    - "README.md"
    - "CLAUDE.md"

repositories:
  app-ios:
    path: "packages/app-ios"
    url: "git@github.com:org/app-ios.git"
    branch: "develop"
    role: "iOS Client App (Swift / SwiftUI)"
    tags: ["frontend", "ios"]

  app-backend:
    path: "packages/app-backend"
    url: "git@github.com:org/app-backend.git"
    role: "Core API Server (Go / gRPC)"
    tags: ["backend", "api"]
    sync: false   # optional: exclude from kusabi sync
```

リポジトリフィールド:

| フィールド | 意味 |
| :--- | :--- |
| `branch` | clone 時および sync 時に追跡する branch（省略時は remote の default） |
| `sync: false` | `kusabi sync` の対象外（`exec` / `context` / `status` は引き続き対象） |
| `includes` | 子ごとの context 観測パス（省略時は親の `context.includes` を継承） |

## CLI

バイナリ名: `kusabi`（短縮: `ksb`）。`git-kusabi` を PATH に置くと `git kusabi` としても呼べる。

| コマンド | 副作用 |
| :--- | :--- |
| `kusabi init [--force]` | `kusabi.yaml`・`AGENTS.md` テンプレート生成。`.gitignore` に除外ブロック追加 |
| `kusabi add <name> <url> [--path] [--role] [--tags] [--branch]` | 宣言更新。該当パスを `.gitignore` に追加 |
| `kusabi remove <name>` | 宣言から削除。`.gitignore` から該当パスを除去（ローカル実体は削除しない） |
| `kusabi sync [--depth=N]` | 未クローンを clone、既存を宣言 branch へ整列して pull。dirty / 未宣言の detached HEAD は skip（warn）。pull 失敗は fail |
| `kusabi status [--json]` | 各子のブランチ・作業ツリー状態を表示（worktree は `[worktree]` / `is_worktree`） |
| `kusabi exec [--tag=T] [--repo=N]... [--skip-uncloned] "<command>"` | 宣言対象（タグ・名前で絞り込み可）でシェルコマンドを並列実行 |
| `kusabi context [--tree] [--json]` | 宣言・子文書を観測して STDOUT 出力。リポジトリは変更しない |
| `kusabi doctor [--migrate-gitignore] [--fix-remote]` | 宣言・除外・clone・branch・remote の不整合を検出。問題があれば非ゼロ終了 |

いずれの作用コマンドも、**失敗**（`Err`）が1件でもあれば非ゼロ終了する。**skip**（dirty / detached / sync 無効など）は warn 表示のみで終了コード 0。

## `kusabi context` 出力

観測のみ。宣言にも子にも無い方針文は付けない。

Markdown 出力の構成:

1. **Meta Architecture & Global Policy** — `context.agents` が指すファイルの内容（存在する場合）
2. **Parent Context Files** — `context.paths` に列挙された親側ファイルを宣言順に転載（1件以上宣言がある場合）
3. **Bound Repositories Overview** — 宣言上の名前・パス・役割・タグ
4. **Repository Contexts** — 各子について、`context.includes` に列挙されたパスのうち**存在するものをすべて**順に転載。`--tree` 時はディレクトリ構造を付加

JSON 出力（`--json`）は同内容を構造化する。`includes` の各ファイルは配列要素として列挙し、`context.paths` は `meta.parent_context_files` に列挙する。

欠落した include は捏造せず、該当ファイルを省略する。いずれの include も存在しない子についてのみ、観測結果として欠落を一文で示す。`context.paths` の欠落は宣言の明示的な対象であるため、省略ではなく欠落として示す（Markdown は `_(missing: path)_`、JSON は `missing: true`）。
