// MongoDB initialization script for audit database
db = db.getSiblingDB('audit_db');

// Create user for the audit service
db.createUser({
  user: 'audit_user',
  pwd: 'audit_password',
  roles: [
    {
      role: 'readWrite',
      db: 'audit_db'
    }
  ]
});

// Create indexes for better query performance
db.logs.createIndex({ "userId": 1 });
db.logs.createIndex({ "action": 1 });
db.logs.createIndex({ "timestamp": -1 });
db.logs.createIndex({ "userId": 1, "timestamp": -1 });
db.logs.createIndex({ "action": 1, "timestamp": -1 });

// Create a compound index for analytics queries
db.logs.createIndex({ 
  "userId": 1, 
  "action": 1, 
  "timestamp": -1 
});

// Create text index for searching through metadata
db.logs.createIndex({ 
  "action": "text", 
  "metadata": "text" 
});

print('MongoDB audit database initialized successfully!');
