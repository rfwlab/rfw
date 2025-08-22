FROM golang:1.24.3

RUN apt-get update && apt-get install -y \
    git \
    curl \
    && rm -rf /var/lib/apt/lists/*

RUN curl -L https://github.com/rfwlab/rfw/releases/download/continuous/rfw -o /usr/local/bin/rfw && \
    chmod +x /usr/local/bin/rfw

RUN curl -L https://github.com/tailwindlabs/tailwindcss/releases/download/v4.1.12/tailwindcss-linux-arm64 -o /usr/local/bin/tailwindcss && \
    chmod +x /usr/local/bin/tailwindcss

WORKDIR /app

RUN git clone https://github.com/rfwlab/rfw.git

WORKDIR /app/rfw/docs

EXPOSE 8080

CMD ["rfw", "dev"]
