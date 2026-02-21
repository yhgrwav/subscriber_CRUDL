CREATE TABLE IF NOT EXISTS subscriptions(
                                            ID SERIAL PRIMARY KEY,
                                            service_name VARCHAR NOT NULL,
                                            price INTEGER NOT NULL,
                                            user_id uuid NOT NULL,
    --т.к. в примере используется "месяц-год", то я не буду тянуть timestamp,
    --который вдвое больше байтов занимает, а оставлю тип DATE, end_time оставлю необязательным
    --на случай, когда подписка куплена, но не активирована,
    --но в целом эти тонкости можно было раздуть еще больше (вне рамок тестового на неделю)
                                            start_time DATE NOT NULL,
                                            end_time DATE
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);