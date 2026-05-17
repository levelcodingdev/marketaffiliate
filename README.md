# FRONTEND
1. npm install
2. npm run dev 

# BACKEND ( in the root of the project )
1. docker compose up -d postgres
2. docker compose up --build backend ( THIS STARTS THE PROJECT )
3. docker compose run --rm backend /app/migrate ( migrate once )
4. docker compose run --rm backend /app/seed ( seed once )

## Stripe Setup ( YOU DO THIS EVERY TIME YOU START PROJECT )
Setup env in backend:

STRIPE_SECRET_KEY=
STRIPE_WEBHOOK_SECRET=
STRIPE_CURRENCY=usd

1. docker compose --profile stripe up stripe
2. Copy ```whsec_``` to STRIPE_WEBHOOK_SECRET env