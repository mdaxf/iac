// IAC MongoDB Initialization Script

// Switch to IAC database
db = db.getSiblingDB('iac');

// Create user for the database
db.createUser({
  user: 'iac_user',
  pwd: 'iac_pass',
  roles: [
    { role: 'readWrite', db: 'iac' },
    { role: 'dbAdmin', db: 'iac' }
  ]
});

// Create collections
db.createCollection('users');
db.createCollection('sessions');
db.createCollection('audit_log');
db.createCollection('configurations');

// Create indexes for users collection
db.users.createIndex({ "username": 1 }, { unique: true });
db.users.createIndex({ "email": 1 }, { unique: true });
db.users.createIndex({ "status": 1 });
db.users.createIndex({ "created_at": -1 });

// Create indexes for sessions collection
db.sessions.createIndex({ "session_id": 1 }, { unique: true });
db.sessions.createIndex({ "user_id": 1 });
db.sessions.createIndex({ "last_activity": -1 });
db.sessions.createIndex({ "created_at": 1 }, { expireAfterSeconds: 86400 }); // TTL index: 24 hours

// Create indexes for audit_log collection
db.audit_log.createIndex({ "user_id": 1 });
db.audit_log.createIndex({ "action": 1 });
db.audit_log.createIndex({ "created_at": -1 });
db.audit_log.createIndex({ "entity_type": 1, "entity_id": 1 });

// Insert sample data
db.users.insertMany([
  {
    uuid: '550e8400-e29b-41d4-a716-446655440000',
    username: 'admin',
    email: 'admin@iac.local',
    password_hash: '$2a$10$rZfE8qvd1xqY.T9hG3V8H.',
    first_name: 'Admin',
    last_name: 'User',
    status: 'active',
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    uuid: '660e8400-e29b-41d4-a716-446655440001',
    username: 'testuser',
    email: 'test@iac.local',
    password_hash: '$2a$10$rZfE8qvd1xqY.T9hG3V8H.',
    first_name: 'Test',
    last_name: 'User',
    status: 'active',
    created_at: new Date(),
    updated_at: new Date()
  }
]);

// Insert sample configuration
db.configurations.insertOne({
  name: 'default',
  description: 'Default IAC configuration',
  settings: {
    max_connections: 100,
    timeout: 30,
    enable_logging: true
  },
  created_at: new Date(),
  updated_at: new Date()
});

print('IAC MongoDB database initialized successfully');
