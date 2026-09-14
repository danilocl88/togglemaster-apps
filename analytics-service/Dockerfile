FROM python:3.11-slim

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

WORKDIR /app

RUN groupadd --system app     && useradd --system --gid app --home-dir /app app

COPY requirements.txt .

RUN pip install     --no-cache-dir     --disable-pip-version-check     -r requirements.txt

COPY . .

RUN chown -R app:app /app

USER app

EXPOSE 8005

CMD ["python", "app.py"]
