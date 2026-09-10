INSERT INTO users (email, password_hash) VALUES
('eduard@example.com', '$2a$10$LS6erZQNlNsz9GRyZhHed.J0xyHUsUXEMnWHMitMETW/x1.JMlQQi'),
('bob@example.com', '$2a$10$LS6erZQNlNsz9GRyZhHed.J0xyHUsUXEMnWHMitMETW/x1.JMlQQi'),
('charlie@example.com', '$2a$10$LS6erZQNlNsz9GRyZhHed.J0xyHUsUXEMnWHMitMETW/x1.JMlQQi');

INSERT INTO tasks (title, description, deadline, status, creator_id, assignee_id, created_at)
SELECT 
    'Test Task #' || i AS title,
    
    'Description for task #' || i AS description,
    
    NOW() + (random() * 20 - 5) * INTERVAL '1 day' AS deadline,
    
    (ARRAY['created', 'in_progress', 'completed', 'overdue'])[mod(i, 4) + 1]::task_status AS status,
    
    (SELECT id FROM users ORDER BY id LIMIT 1 OFFSET floor(random() * 3)) AS creator_id,
    (SELECT id FROM users ORDER BY id LIMIT 1 OFFSET floor(random() * 3)) AS assignee_id,
    
    NOW() - (random() * 10) * INTERVAL '1 day' AS created_at
FROM generate_series(1, 15) AS i;
