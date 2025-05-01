FROM python:3.13-alpine

WORKDIR /app

COPY requirements.txt .
RUN pip install --upgrade pip && pip install -r requirements.txt

COPY . .

ENV PYTHONPATH=/app
CMD ["pytest", "-vv", "-ra", "-s"]
