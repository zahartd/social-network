FROM python:3.13-alpine

WORKDIR /app

COPY requirements.txt .
RUN pip install --upgrade pip && pip install -r requirements.txt

COPY . .

CMD ["pytest", "--maxfail=1", "--disable-warnings", "-q"]
