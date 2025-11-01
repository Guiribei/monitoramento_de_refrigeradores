# Refrigerator Monitor

<p align="center">
    <img src="https://github.com/Guiribei/monitoramento_de_refrigeradores/blob/main/snowflake.svg" />
</p>

Collects data from a smart plug. A React frontend that communicates with a Go backend API.  

---
.env example:
```bash
PORT=XXXX
ALLOWED_ORIGIN=XXXX


RATE_WINDOW_SECONDS=XXXX
TUYA_BASE_URL=XXXX
TUYA_DEVICE_ID=XXXX

TUYA_CLIENT_ID=XXXX
TUYA_CLIENT_SECRET=XXXX
TUYA_ACCESS_TOKEN=XXXX

DATA_DIR=XXXX
```
---

## Run project locally

Go to frontend folder
```bash
cd frontend
```

Install frontend dependencies:
```bash
npm install
```

Run development server:
```bash
npm run dev
```
#### In a separated terminal:
Go to backend folder
```bash
cd backend
```

Run development backend server:
```bash
go run .
```

---


## Technologies

- [Vite](https://vitejs.dev/)
- [Go](https://go.dev/)

---

## License

For internal or academic use only. Redistribution is not allowed without permission.
