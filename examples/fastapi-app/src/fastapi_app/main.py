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
    uvicorn.run("fastapi_app.main:app", host="0.0.0.0", port=8000, reload=True)


if __name__ == "__main__":
    serve()
