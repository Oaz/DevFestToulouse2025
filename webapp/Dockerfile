FROM node:20-alpine

WORKDIR /app

COPY game/rules.yaml ../game/rules.yaml

# Copier uniquement package.json et package-lock.json pour profiter du cache
COPY webapp/package*.json ./

# Installer toutes les dépendances (prod + dev)
RUN npm install

# Copier le reste de l'application
COPY webapp/ ./

# Exposer le port de Vite
EXPOSE 5173

# CMD dev via npm run (Vite sera trouvé)
CMD ["npm", "run", "dev"]