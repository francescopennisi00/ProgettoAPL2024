FROM python:3.11

WORKDIR /app
ENV PYTHONPATH "/app"
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

COPY worker ./worker

CMD [ "python", "worker/worker_main.py" ]