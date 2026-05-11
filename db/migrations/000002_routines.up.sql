CREATE OR REPLACE FUNCTION get_user_by_email(p_email TEXT)
RETURNS SETOF users AS $$
    SELECT * FROM users WHERE email = p_email;
$$ LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION get_refresh_token(p_token TEXT)
RETURNS SETOF refresh_tokens AS $$
    SELECT * FROM refresh_tokens WHERE token = p_token AND expires_at > NOW();
$$ LANGUAGE sql STABLE;

CREATE OR REPLACE PROCEDURE save_refresh_token(
    p_user_id   UUID,
    p_token     TEXT,
    p_expires_at TIMESTAMPTZ
) AS $$
    INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES (p_user_id, p_token, p_expires_at);
$$ LANGUAGE sql;

CREATE OR REPLACE PROCEDURE delete_refresh_token(p_token TEXT) AS $$
    DELETE FROM refresh_tokens WHERE token = p_token;
$$ LANGUAGE sql;

CREATE OR REPLACE PROCEDURE save_qr_computation(
    p_user_id UUID,
    p_input   JSONB,
    p_q       JSONB,
    p_r       JSONB,
    p_success BOOLEAN,
    p_error   TEXT
) AS $$
    INSERT INTO qr_computations (user_id, input_matrix, q_matrix, r_matrix, success, error_msg)
    VALUES (p_user_id, p_input, p_q, p_r, p_success, p_error);
$$ LANGUAGE sql;

CREATE OR REPLACE PROCEDURE save_stats_computation(
    p_user_id UUID,
    p_q       JSONB,
    p_r       JSONB,
    p_max     NUMERIC,
    p_min     NUMERIC,
    p_avg     NUMERIC,
    p_sum     NUMERIC,
    p_q_diag  BOOLEAN,
    p_r_diag  BOOLEAN,
    p_success BOOLEAN,
    p_error   TEXT
) AS $$
    INSERT INTO stats_computations
        (user_id, q_matrix, r_matrix, max_value, min_value, avg_value, total_sum, q_diagonal, r_diagonal, success, error_msg)
    VALUES
        (p_user_id, p_q, p_r, p_max, p_min, p_avg, p_sum, p_q_diag, p_r_diag, p_success, p_error);
$$ LANGUAGE sql;
