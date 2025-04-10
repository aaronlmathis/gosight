-- Users table
users (
    id UUID PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    mfa_secret TEXT,
    created_at TIMESTAMP DEFAULT now()
)

-- Roles table
roles (
    id UUID PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT
)

-- Permissions table
permissions (
    id UUID PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,  -- e.g., "metrics:view", "agents:manage"
    description TEXT
)

-- Join table: user → role (many-to-many)
user_roles (
    user_id UUID REFERENCES users(id),
    role_id UUID REFERENCES roles(id),
    PRIMARY KEY (user_id, role_id)
)

-- Join table: role → permission (many-to-many)
role_permissions (
    role_id UUID REFERENCES roles(id),
    permission_id UUID REFERENCES permissions(id),
    PRIMARY KEY (role_id, permission_id)
)
