# ROPUS — Analyst Control Plane & Demo UI

Frontend control plane and interactive scenario evaluator for **ROPUS (Track 02: AI Risk Manager)**.

For the full system architecture, Go backend orchestrator, and ML evaluation results, refer to the [Root README](../README.md).

---

## ⚡ Local Development

### Native Setup:
```bash
cd frontend
npm install
npm run dev
```

The application will be available at [http://localhost:3000](http://localhost:3000).

### Stack Setup (via Docker):
From the repository root:
```bash
make up
```

---

## 🛠️ Verification Commands
```bash
cd frontend
npm run lint
npm run build
```
