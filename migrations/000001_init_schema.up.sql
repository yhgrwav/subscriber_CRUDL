CREATE TABLE IF NOT EXISTS subscriptions(
                                            ID SERIAL PRIMARY KEY,
                                            service_name VARCHAR NOT NULL,
                                            price INTEGER NOT NULL,
                                            user_id uuid NOT NULL,
                                            start_date DATE NOT NULL,
                                            end_date DATE
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
