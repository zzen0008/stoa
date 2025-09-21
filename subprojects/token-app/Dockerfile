# Use an official Python runtime as a parent image
FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

RUN mkdir -p /app/.streamlit
COPY .streamlit/secrets.toml /app/.streamlit/
COPY app.py /app/

EXPOSE 8501

CMD ["streamlit", "run", "app.py"]
