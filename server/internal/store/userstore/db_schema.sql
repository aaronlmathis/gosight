CREATE TABLE users (
    id UUID PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    first_name TEXT,
    last_name TEXT,
    password_hash TEXT NOT NULL,
    mfa_secret TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    mfa_enabled BOOLEAN DEFAULT FALSE,
    mfa_method TEXT,
    webauthn_creds BYTEA,
    sso_provider TEXT,
    sso_id TEXT,
    username TEXT UNIQUE,
    last_login TIMESTAMP
);

-- Optional: index to enforce unique SSO identity
CREATE UNIQUE INDEX unique_sso_identity ON users (sso_provider, sso_id);
CREATE TABLE roles (
    id UUID PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT
);

CREATE TABLE permissions (
    id UUID PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT
);

CREATE TABLE user_roles (
    user_id UUID NOT NULL,
    role_id UUID NOT NULL,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

CREATE TABLE role_permissions (
    role_id UUID NOT NULL,
    permission_id UUID NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);
