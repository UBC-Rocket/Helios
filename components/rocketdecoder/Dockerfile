FROM python:3.13

ENV PYTHONUNBUFFERED=1

# Install uv from prebuilt binary
COPY --from=ghcr.io/astral-sh/uv:0.9.2 /uv /uvx /bin/

WORKDIR /
COPY . .

RUN uv sync --locked

EXPOSE 5000

CMD ["uv", "run", "main.py", "--host", "0.0.0.0", "--port", "5000"]