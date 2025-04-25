# gh2bb — GitHub to Bitbucket SSH Mirror Migrator

**gh2bb** is a lightweight, high‑fidelity CLI tool written in Go that automates the migration of a GitHub repository to Bitbucket using SSH. It performs a bare mirror‑clone of your source repo and mirror‑pushes all refs (branches, tags) to your destination repo, preserving history and metadata.

---

## 🔑 Features

- **SSH‑based migration**: No HTTP tokens needed once SSH keys are configured.
- **Mirror clone & push**: Uses `git clone --mirror` and `git push --mirror` for complete fidelity.
- **Env‑driven config**: Reads `GH_PREFIX` from a `.env` file or environment.
- **Dry‑run mode**: Preview the steps without executing any Git commands.
- **Zero‑dependency runtime**: Requires Go 1.20+ (static binary) and `git` CLI.

---

## 🚀 Quick Start

1. **Prepare your `.env`**

   ```bash
   # .env
   GH_PREFIX={your_github_user}
   ```

2. **Build**

   ```bash
   go build -o gh2bb
   ```

3. **Run (dry‑run)**

   ```bash
   ./gh2bb \
     --dest-url git@bitbucket.org:<workspace>/<repo-name>.git \
     --dry-run
   ```

4. **Migrate**

   ```bash
   ./gh2bb \
     --dest-url git@bitbucket.org:<workspace>/<repo-name>.git
   ```

---

## 🎛 Configuration

| Variable   | Description                                               | Default        |
|------------|-----------------------------------------------------------|----------------|
| `GH_PREFIX`| GitHub organization or user name (prefix for source URL)  | **required**   |

- The tool reads `GH_PREFIX` at startup. If unset, it will error out.

---

## 📖 Usage

```text
Usage:
  gh2bb [flags]

Flags:
  --dest-url string   SSH URL of Bitbucket repo (e.g. git@bitbucket.org:ws/repo.git) (required)
  --dry-run           Print Git commands without executing
  -h, --help          help for gh2bb
```

1. **dest-url**: Fully qualified SSH path to your target Bitbucket repo. The tool will extract the repo name from this URL.
2. **dry-run**: Prints each Git invocation instead of running it.

---

## 🧩 How It Works

1. **Derive source URL** from `GH_PREFIX` and repo name:
   ```none
   sourceURL = "git@github.com:" + GH_PREFIX + "/<repo>.git"
   ```
2. **Clone bare mirror**:
   ```bash
   git clone --mirror "$sourceURL" /tmp/gh2bb-<random>/repo.git
   ```
3. **Set push URL** on `origin` to your Bitbucket destination:
   ```bash
   git -C /tmp/.../repo.git remote set-url --push origin "$destURL"
   ```
4. **Mirror‑push** all refs:
   ```bash
   git -C /tmp/.../repo.git push --mirror origin
   ```

---

## ⚙️ Prerequisites

- Go 1.20+ installed (for building)
- `git` CLI available in your `PATH`
- SSH keys configured for both GitHub and Bitbucket

---

## 📜 License

This project is licensed under the [MIT License](LICENSE).

---

_Enjoy seamless GitHub → Bitbucket migrations!_

