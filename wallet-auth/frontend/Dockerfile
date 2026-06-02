FROM node:20-alpine AS builder

WORKDIR /app

ARG VITE_API_URL=http://localhost:8080
ARG VITE_WALLETCONNECT_PROJECT_ID

ENV VITE_API_URL=$VITE_API_URL
ENV VITE_WALLETCONNECT_PROJECT_ID=$VITE_WALLETCONNECT_PROJECT_ID

COPY package.json ./
RUN npm install

COPY . .
RUN npm run build

FROM nginx:1.27-alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

USER nginx
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
