import os
from typing import Union

from fastapi import FastAPI
import uvicorn

app = FastAPI()


@app.get("/")
def read_root():
    return {"Hello": "World"}


@app.get("/items/{item_id}")
def read_item(item_id: int, q: Union[str, None] = None):
    return {"item_id": item_id, "q": q}


def serve():
    port = os.getenv("PORT", 8080)
    log_level = os.getenv("LOG_LEVEL", "info")
    workers = os.getenv("WORKERS", 1)
    print("Starting FastAPI server...")
    print(f"Port: {port}")
    print(f"Log Level: {log_level}")
    print(f"Workers: {workers}")
    uvicorn.run(
        "fastapi_app.main:app",
        host="0.0.0.0",
        port=int(port),
        log_level=log_level,
        workers=int(workers),
    )


if __name__ == "__main__":
    serve()
