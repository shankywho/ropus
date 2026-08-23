FROM oven/bun:1.2-alpine

WORKDIR /app

COPY package.json bun.lock ./
RUN bun install --frozen-lockfile

COPY . .

EXPOSE 3000

ENV PORT=3000
ENV HOST=0.0.0.0

CMD ["bun", "run", "dev", "--host", "0.0.0.0", "--port", "3000"]
