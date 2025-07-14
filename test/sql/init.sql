-- Test schema for dbutil-gen code generator
-- This schema covers all PostgreSQL data types and edge cases

-- Enable UUID extension for UUID support
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create a simple UUID v7-like function for testing
-- Note: This is a simplified version for testing purposes
-- In production, you would use a proper UUID v7 implementation
CREATE OR REPLACE FUNCTION uuid_generate_v7() RETURNS UUID AS $$
BEGIN
    -- Generate a UUID v4 for testing (in production, use proper UUID v7)
    RETURN uuid_generate_v4();
END;
$$ LANGUAGE plpgsql;

-- Users table - Basic table with UUID primary key (UUID v7 required)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login TIMESTAMP WITH TIME ZONE,
    metadata JSONB,
    age INTEGER,
    balance DECIMAL(10,2),
    profile_picture_url TEXT
);

-- Profiles table - One-to-one relationship with users
CREATE TABLE profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bio TEXT,
    avatar_url TEXT,
    website_url TEXT,
    location VARCHAR(255),
    birth_date DATE,
    phone VARCHAR(20),
    preferences JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Posts table - One-to-many relationship with users
CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    content TEXT NOT NULL,
    excerpt TEXT,
    slug VARCHAR(500) UNIQUE NOT NULL,
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    published_at TIMESTAMP WITH TIME ZONE,
    view_count INTEGER DEFAULT 0,
    like_count INTEGER DEFAULT 0,
    tags TEXT[] DEFAULT '{}',
    featured_image_url TEXT,
    seo_data JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Comments table - Hierarchical/tree structure with self-referencing FK
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    is_approved BOOLEAN DEFAULT false,
    upvotes INTEGER DEFAULT 0,
    downvotes INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Categories table - Many-to-many relationship with posts (via post_categories)
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    slug VARCHAR(100) UNIQUE NOT NULL,
    parent_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Post-Categories junction table - Many-to-many relationship
CREATE TABLE post_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(post_id, category_id)
);

-- Files table - Binary data and file metadata
CREATE TABLE files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    file_hash VARCHAR(64) NOT NULL,
    storage_path TEXT NOT NULL,
    is_public BOOLEAN DEFAULT false,
    download_count INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Comprehensive data types table - Tests all PostgreSQL types
CREATE TABLE data_types_test (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    
    -- String types
    text_field TEXT,
    varchar_field VARCHAR(255),
    char_field CHAR(10),
    
    -- Numeric types
    smallint_field SMALLINT,
    integer_field INTEGER,
    bigint_field BIGINT,
    decimal_field DECIMAL(10,2),
    numeric_field NUMERIC(15,5),
    real_field REAL,
    double_field DOUBLE PRECISION,
    
    -- Boolean
    boolean_field BOOLEAN,
    
    -- Date/Time types
    date_field DATE,
    time_field TIME,
    timestamp_field TIMESTAMP,
    timestamptz_field TIMESTAMP WITH TIME ZONE,
    interval_field INTERVAL,
    
    -- UUID
    uuid_field UUID,
    
    -- JSON types
    json_field JSON,
    jsonb_field JSONB,
    
    -- Array types
    text_array_field TEXT[],
    integer_array_field INTEGER[],
    uuid_array_field UUID[],
    
    -- Network types
    inet_field INET,
    cidr_field CIDR,
    macaddr_field MACADDR,
    
    -- Other types
    bytea_field BYTEA,
    xml_field XML,
    
    -- Nullable versions (to test pgtype integration)
    nullable_text TEXT,
    nullable_integer INTEGER,
    nullable_boolean BOOLEAN,
    nullable_timestamp TIMESTAMP WITH TIME ZONE,
    nullable_uuid UUID,
    nullable_jsonb JSONB,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Test table with non-UUID primary key (should be rejected by generator)
CREATE TABLE invalid_pk_table (
    id SERIAL PRIMARY KEY,  -- This should cause dbutil-gen to reject the table
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Test table with composite primary key (should be rejected by generator)
CREATE TABLE composite_pk_table (
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (tenant_id, user_id)
);

-- Indexes for performance testing
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_active_created ON users(is_active, created_at);
CREATE INDEX idx_posts_user_status ON posts(user_id, status);
CREATE INDEX idx_posts_published ON posts(published_at) WHERE status = 'published';
CREATE INDEX idx_comments_post_parent ON comments(post_id, parent_id);
CREATE INDEX idx_profiles_user ON profiles(user_id);
CREATE UNIQUE INDEX idx_profiles_user_unique ON profiles(user_id);

-- Insert test data
INSERT INTO users (id, name, email, password_hash, is_active, age, balance, metadata) VALUES
    (uuid_generate_v7(), 'John Doe', 'john@example.com', 'hashed_password_1', true, 30, 1000.50, '{"role": "admin", "preferences": {"theme": "dark"}}'),
    (uuid_generate_v7(), 'Jane Smith', 'jane@example.com', 'hashed_password_2', true, 25, 750.25, '{"role": "user", "preferences": {"theme": "light"}}'),
    (uuid_generate_v7(), 'Bob Johnson', 'bob@example.com', 'hashed_password_3', false, 35, 0.00, '{"role": "user", "suspended": true}'),
    (uuid_generate_v7(), 'Alice Brown', 'alice@example.com', 'hashed_password_4', true, 28, 2500.75, '{"role": "moderator", "verified": true}');

-- Insert profiles for users
INSERT INTO profiles (id, user_id, bio, location, preferences)
SELECT 
    uuid_generate_v7(),
    u.id,
    'Bio for ' || u.name,
    CASE 
        WHEN u.name = 'John Doe' THEN 'New York, NY'
        WHEN u.name = 'Jane Smith' THEN 'San Francisco, CA'
        WHEN u.name = 'Bob Johnson' THEN 'Chicago, IL'
        ELSE 'Boston, MA'
    END,
    '{"notifications": true, "privacy": "public"}'
FROM users u;

-- Insert categories
INSERT INTO categories (id, name, description, slug, sort_order) VALUES
    (uuid_generate_v7(), 'Technology', 'Tech-related posts', 'technology', 1),
    (uuid_generate_v7(), 'Programming', 'Programming tutorials and tips', 'programming', 2),
    (uuid_generate_v7(), 'Web Development', 'Web development topics', 'web-development', 3),
    (uuid_generate_v7(), 'Database', 'Database design and optimization', 'database', 4);

-- Insert posts
INSERT INTO posts (id, user_id, title, content, excerpt, slug, status, published_at, view_count, like_count, tags)
SELECT 
    uuid_generate_v7(),
    u.id,
    'Sample Post by ' || u.name,
    'This is the content of a sample post by ' || u.name || '. It contains multiple paragraphs and demonstrates the text field capabilities.',
    'Sample excerpt for post by ' || u.name,
    'sample-post-' || LOWER(REPLACE(u.name, ' ', '-')),
    CASE WHEN u.is_active THEN 'published' ELSE 'draft' END,
    CASE WHEN u.is_active THEN NOW() - INTERVAL '1 day' ELSE NULL END,
    FLOOR(RANDOM() * 1000)::INTEGER,
    FLOOR(RANDOM() * 100)::INTEGER,
    ARRAY['sample', 'test', 'demo']
FROM users u;

-- Insert comments
INSERT INTO comments (id, post_id, user_id, content, is_approved, upvotes)
SELECT 
    uuid_generate_v7(),
    p.id,
    u.id,
    'This is a comment on ' || p.title || ' by ' || u.name,
    true,
    FLOOR(RANDOM() * 20)::INTEGER
FROM posts p
CROSS JOIN users u
WHERE u.is_active = true
LIMIT 10;

-- Insert post-category relationships
INSERT INTO post_categories (id, post_id, category_id)
SELECT 
    uuid_generate_v7(),
    p.id,
    c.id
FROM posts p
CROSS JOIN categories c
WHERE RANDOM() < 0.5  -- Randomly assign categories
LIMIT 8;

-- Insert test data for data_types_test table
INSERT INTO data_types_test (
    id, text_field, varchar_field, char_field, smallint_field, integer_field, bigint_field,
    decimal_field, numeric_field, real_field, double_field, boolean_field, date_field,
    time_field, timestamp_field, timestamptz_field, interval_field, uuid_field,
    json_field, jsonb_field, text_array_field, integer_array_field, uuid_array_field,
    inet_field, cidr_field, macaddr_field, nullable_text, nullable_integer, nullable_boolean
) VALUES (
    uuid_generate_v7(),
    'Sample text field with unicode: 你好世界 🌍',
    'VARCHAR field',
    'CHAR      ',  -- Note: CHAR(10) pads with spaces
    32767,
    2147483647,
    9223372036854775807,
    12345.67,
    123456.78901,
    3.14159,
    2.718281828459045,
    true,
    '2024-01-15',
    '14:30:00',
    '2024-01-15 14:30:00',
    '2024-01-15 14:30:00+00',
    '1 year 2 months 3 days 4 hours 5 minutes 6 seconds',
    uuid_generate_v7(),
    '{"key": "value", "number": 42}',
    '{"nested": {"array": [1, 2, 3]}, "boolean": true}',
    ARRAY['one', 'two', 'three'],
    ARRAY[1, 2, 3, 4, 5],
    ARRAY[uuid_generate_v7(), uuid_generate_v7()],
    '192.168.1.1',
    '192.168.1.0/24',
    '08:00:2b:01:02:03',
    'Nullable text value',
    42,
    false
);

-- Insert a row with mostly NULL values to test nullable handling
INSERT INTO data_types_test (id, text_field, nullable_text) VALUES (
    uuid_generate_v7(),
    'Row with mostly NULL values',
    NULL
);

-- Create some views for testing (if generator supports views in the future)
CREATE VIEW active_users_view AS
SELECT 
    u.id,
    u.name,
    u.email,
    u.created_at,
    p.bio,
    COUNT(po.id) as post_count
FROM users u
LEFT JOIN profiles p ON u.id = p.user_id
LEFT JOIN posts po ON u.id = po.user_id
WHERE u.is_active = true
GROUP BY u.id, u.name, u.email, u.created_at, p.bio;

-- Create a function for testing (if generator supports functions in the future)
CREATE OR REPLACE FUNCTION get_user_post_count(user_uuid UUID)
RETURNS INTEGER AS $$
BEGIN
    RETURN (
        SELECT COUNT(*)::INTEGER
        FROM posts
        WHERE user_id = user_uuid
    );
END;
$$ LANGUAGE plpgsql;

-- Add some constraints for testing
ALTER TABLE users ADD CONSTRAINT chk_users_age CHECK (age >= 0 AND age <= 150);
ALTER TABLE posts ADD CONSTRAINT chk_posts_counts CHECK (view_count >= 0 AND like_count >= 0);
ALTER TABLE comments ADD CONSTRAINT chk_comments_votes CHECK (upvotes >= 0 AND downvotes >= 0);

-- Final verification queries (these will run during initialization)
SELECT 'Database initialization completed successfully' as status;
SELECT 'Total users: ' || COUNT(*)::text as user_count FROM users;
SELECT 'Total posts: ' || COUNT(*)::text as post_count FROM posts;
SELECT 'Total comments: ' || COUNT(*)::text as comment_count FROM comments;
SELECT 'Total categories: ' || COUNT(*)::text as category_count FROM categories; 