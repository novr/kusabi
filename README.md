# Kusabi（楔）

*Connect Repositories. Unify Contexts.*

複数の Git リポジトリを束ね、宣言と各子の文書を**観測として**一つの文脈に差し出す CLI。親の Git 履歴に子の実体を混ぜず、Submodule の摩擦も引き受けない。

## インストール

### Homebrew（推奨）

```bash
brew tap novr/taps
brew install kusabi
```

`kusabi` が入る。リリース tarball には `ksb` / `git-kusabi` も同梱される（Formula の `install` は初回作成後に tap 側で3バイナリへ拡張可能）。

### Go install

```bash
go install github.com/novr/kusabi/cmd/kusabi@latest
```

短縮コマンド `ksb`、Git 連携 `git kusabi` も個別に入れられる。

```bash
go install github.com/novr/kusabi/cmd/ksb@latest
go install github.com/novr/kusabi/cmd/git-kusabi@latest
```

## リリース

`v*` タグを push すると GitHub Release と [homebrew-taps](https://github.com/novr/homebrew-taps) の Formula が更新される。

```bash
git tag v0.1.0
git push origin v0.1.0
```

手動実行は Actions の **Release Assets** ワークフローから。

**前提:** リポジトリ secrets に `NOVRD_BOT_CLIENT_ID` / `NOVRD_BOT_KEY` が必要。org 設定で `novr/homebrew-taps` の reusable workflow へのアクセスも有効にすること（[new-tool.md](https://github.com/novr/homebrew-taps/blob/main/docs/new-tool.md) 参照）。

## クイックスタート

```bash
mkdir my-meta && cd my-meta
git init

kusabi init
kusabi add app-ios git@github.com:org/app-ios.git --role "iOS App"
kusabi add app-backend git@github.com:org/app-backend.git --role "API Server"
kusabi sync
kusabi doctor
kusabi context | pbcopy
```

## コマンド

| コマンド | 説明 |
| :--- | :--- |
| `init` | `kusabi.yaml`・`AGENTS.md`・`.gitignore` を初期化 |
| `add` / `remove` | 宣言の追加・削除（`--branch` 対応） |
| `sync` | 子の clone / 宣言 branch への整列 / pull（dirty・未宣言 detached は skip） |
| `status` | 各子のブランチ・作業ツリー状態（`--json` で機械可読） |
| `exec` | 宣言対象への一括実行（`--repo` / `--tag` / `--skip-uncloned`） |
| `context` | 宣言・子文書を観測して出力 |
| `doctor` | 宣言・clone・branch・remote の検査（`--fix-remote` / `--migrate-gitignore`） |

詳細は [PRODUCT.md](PRODUCT.md)。設計上の境界は [AGENTS.md](AGENTS.md)。

## 開発

```bash
go test ./...
go run ./cmd/kusabi --help
```

## ライセンス

[MIT](LICENSE)
