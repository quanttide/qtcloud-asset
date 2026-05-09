from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware


app = FastAPI(title="QtCloud Asset Provider", version="0.1.0")

PROVIDER_BASE_URL = "https://provide-package-tuzknrkwac.cn-hangzhou.fcapp.run"

app.add_middleware(
    CORSMiddleware,
    allow_origins=[
        "https://asset.quanttide.com",
        "http://localhost:8080",
        "http://localhost:9000",
    ],
    allow_credentials=False,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.get("/health")
def health():
    return {"status": "ok", "service": "qtcloud-asset-provider"}


@app.get("/")
def root():
    return {
        "name": "qtcloud-asset-provider",
        "description": "QtCloud Asset API provider",
        "status": "ready",
    }


@app.get("/config")
def config():
    return {
        "provider_base_url": PROVIDER_BASE_URL,
        "studio_origin": "https://asset.quanttide.com",
        "cors": "enabled",
    }
