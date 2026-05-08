from fastapi import FastAPI


app = FastAPI(title="QtCloud Asset Provider", version="0.1.0")


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
