FROM node:18-alpine

# Set working directory in the container
WORKDIR /usr/src/app

# Copy package.json and package-lock.json to the container
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy the rest of the project files to the container
COPY . .

# Set environment variables for Next.js
ENV NODE_ENV=production

ENV NEXT_PUBLIC_API_URL=http://192.168.23.61:8080

# Build the Next.js project
RUN npm run build

# Expose the port Next.js will run on
EXPOSE 3000

ENV HOSTNAME="0.0.0.0"

# Start the Next.js application
# CMD ["npm", "run", "start"]
CMD ["npm", "start"]

