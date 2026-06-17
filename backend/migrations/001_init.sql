CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone VARCHAR(20) UNIQUE,
    email VARCHAR(255) UNIQUE,
    wechat_open_id VARCHAR(128) UNIQUE,
    wechat_union_id VARCHAR(128),
    password_hash VARCHAR(255) NOT NULL,
    nickname VARCHAR(50) UNIQUE NOT NULL,
    avatar_url VARCHAR(500),
    gender SMALLINT DEFAULT 0,  -- 0: secret, 1: male, 2: female
    birth_date DATE,
    bio VARCHAR(500),
    city VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    last_active_at TIMESTAMPTZ DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_city ON users(city);
CREATE INDEX idx_users_wechat_union ON users(wechat_union_id);
CREATE INDEX idx_users_last_active ON users(last_active_at);

-- Interest tags
CREATE TABLE interest_tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    category VARCHAR(50) NOT NULL,
    icon VARCHAR(100)
);

-- User interests (many-to-many)
CREATE TABLE user_interests (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    tag_id INT REFERENCES interest_tags(id) ON DELETE CASCADE,
    weight INT DEFAULT 1,
    PRIMARY KEY (user_id, tag_id)
);

-- Personality questions
CREATE TABLE personality_questions (
    id SERIAL PRIMARY KEY,
    dimension VARCHAR(50) NOT NULL,  -- openness, conscientiousness, extraversion, agreeableness, neuroticism
    question_text VARCHAR(500) NOT NULL,
    sort_order INT DEFAULT 0
);

-- Personality options for each question
CREATE TABLE personality_options (
    id SERIAL PRIMARY KEY,
    question_id INT REFERENCES personality_questions(id) ON DELETE CASCADE,
    option_text VARCHAR(200) NOT NULL,
    score INT NOT NULL,  -- 1-5 Likert scale
    sort_order INT DEFAULT 0
);

-- User personality answers
CREATE TABLE user_personality_answers (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    question_id INT REFERENCES personality_questions(id) ON DELETE CASCADE,
    option_id INT REFERENCES personality_options(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, question_id)
);

-- Avatar components (virtual avatar system)
CREATE TABLE avatar_components (
    id SERIAL PRIMARY KEY,
    category VARCHAR(50) NOT NULL,  -- hair, eyes, face, clothes, accessory, background
    name VARCHAR(100) NOT NULL,
    image_url VARCHAR(500) NOT NULL,
    rarity SMALLINT DEFAULT 1  -- 1: common, 2: rare, 3: epic
);

-- Matches between users
CREATE TABLE matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id_1 UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_id_2 UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    score FLOAT NOT NULL DEFAULT 0,
    status SMALLINT DEFAULT 0,  -- 0: pending, 1: matched, 2: rejected
    created_at TIMESTAMPTZ DEFAULT NOW(),
    matched_at TIMESTAMPTZ,
    UNIQUE(user_id_1, user_id_2)
);

CREATE INDEX idx_matches_user1 ON matches(user_id_1, status);
CREATE INDEX idx_matches_user2 ON matches(user_id_2, status);

-- Chat rooms (1-on-1)
CREATE TABLE chat_rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id UUID REFERENCES matches(id) ON DELETE CASCADE,
    user_id_1 UUID NOT NULL REFERENCES users(id),
    user_id_2 UUID NOT NULL REFERENCES users(id),
    last_message_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_chat_rooms_user1 ON chat_rooms(user_id_1);
CREATE INDEX idx_chat_rooms_user2 ON chat_rooms(user_id_2);

-- Messages
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES chat_rooms(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id),
    client_msg_id UUID UNIQUE NOT NULL,
    content_type SMALLINT DEFAULT 1,  -- 1: text, 2: image, 3: sticker
    content TEXT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    sent_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_messages_room ON messages(room_id, sent_at);
CREATE INDEX idx_messages_client_id ON messages(client_msg_id);

-- Device tokens for push notifications
CREATE TABLE device_tokens (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(500) NOT NULL,
    platform VARCHAR(20) NOT NULL,  -- ios, android
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, token)
);

-- Seed interest tags (80+ tags across 8 categories)
INSERT INTO interest_tags (name, category, icon) VALUES
-- entertainment
('ACG', 'entertainment', '🎬'),
('电影', 'entertainment', '🎬'),
('VLOG', 'entertainment', '📹'),
('脱口秀', 'entertainment', '🎙️'),
('舞台剧', 'entertainment', '🎪'),
('韩剧', 'entertainment', '📺'),
('美剧', 'entertainment', '📺'),
('日剧', 'entertainment', '📺'),
('纪录片', 'entertainment', '🎥'),
('综艺', 'entertainment', '📺'),
('Livehouse', 'entertainment', '🎶'),
-- music
('K-pop', 'music', '🎵'),
('Hip-hop', 'music', '🎤'),
('摇滚', 'music', '🎸'),
('电子音乐', 'music', '🎧'),
('R&B', 'music', '🎷'),
('爵士', 'music', '🎺'),
('古典乐', 'music', '🎻'),
('独立音乐', 'music', '🎸'),
('华语流行', 'music', '🎤'),
('J-pop', 'music', '🎵'),
('乐器演奏', 'music', '🎹'),
-- gaming
('王者荣耀', 'gaming', '👑'),
('原神', 'gaming', '✨'),
('和平精英', 'gaming', '🔫'),
('LOL', 'gaming', '⚔️'),
('电竞', 'gaming', '🏆'),
('主机游戏', 'gaming', '🎮'),
('独立游戏', 'gaming', '🕹️'),
('桌游', 'gaming', '🎲'),
('崩坏：星穹铁道', 'gaming', '🚂'),
('Dota2', 'gaming', '🛡️'),
('瓦罗兰特', 'gaming', '🎯'),
-- social
('剧本杀', 'social', '🎭'),
('密室逃脱', 'social', '🔐'),
('狼人杀', 'social', '🐺'),
('桌游吧', 'social', '🎲'),
('KTV', 'social', '🎤'),
('酒局', 'social', '🍷'),
('派对', 'social', '🎉'),
('志愿者', 'social', '🤝'),
('语言交换', 'social', '🗣️'),
-- outdoor
('露营', 'outdoor', '⛺'),
('徒步', 'outdoor', '🥾'),
('骑行', 'outdoor', '🚴'),
('自驾游', 'outdoor', '🚗'),
('钓鱼', 'outdoor', '🎣'),
('野餐', 'outdoor', '🧺'),
('冲浪', 'outdoor', '🏄'),
('潜水', 'outdoor', '🤿'),
-- sports
('滑雪', 'sports', '🎿'),
('攀岩', 'sports', '🧗'),
('健身', 'sports', '💪'),
('篮球', 'sports', '🏀'),
('足球', 'sports', '⚽'),
('瑜伽', 'sports', '🧘'),
('跑步', 'sports', '🏃'),
('羽毛球', 'sports', '🏸'),
('网球', 'sports', '🎾'),
('游泳', 'sports', '🏊'),
('拳击', 'sports', '🥊'),
('滑板', 'sports', '🛹'),
-- art
('摄影', 'art', '📷'),
('绘画', 'art', '🎨'),
('手作', 'art', '🧶'),
('舞蹈', 'art', '💃'),
('书法', 'art', '✒️'),
('陶艺', 'art', '🏺'),
('花艺', 'art', '💐'),
('音乐制作', 'art', '🎼'),
('设计', 'art', '✏️'),
-- lifestyle
('咖啡', 'lifestyle', '☕'),
('美食探店', 'lifestyle', '🍜'),
('旅行', 'lifestyle', '✈️'),
('宠物', 'lifestyle', '🐱'),
('读书', 'lifestyle', '📚'),
('穿搭', 'lifestyle', '👔'),
('潮玩盲盒', 'lifestyle', '🎁'),
('调酒', 'lifestyle', '🍸'),
('烹饪', 'lifestyle', '🍳'),
('烘焙', 'lifestyle', '🧁'),
('冥想', 'lifestyle', '🧘'),
('极简主义', 'lifestyle', '🌿'),
('养植物', 'lifestyle', '🪴'),
('数码科技', 'lifestyle', '📱'),
('古着', 'lifestyle', '👗');

-- Seed personality questions (Big Five, simplified)
INSERT INTO personality_questions (dimension, question_text, sort_order) VALUES
('extraversion', '周末晚上你更想：', 1),
('extraversion', '在派对上你通常是：', 2),
('agreeableness', '朋友向你倾诉烦恼时，你：', 3),
('agreeableness', '团队合作中你更看重：', 4),
('conscientiousness', '面对截止日期你通常：', 5),
('conscientiousness', '你的房间/工作区通常是：', 6),
('neuroticism', '遇到突发问题时你：', 7),
('neuroticism', '别人对你的评价让你：', 8),
('openness', '对于新鲜事物你：', 9),
('openness', '旅行时你更喜欢：', 10);

-- Seed options for question 1
INSERT INTO personality_options (question_id, option_text, score, sort_order) VALUES
(1, '和好友出去嗨', 5, 1),
(1, '小范围聚会', 4, 2),
(1, '看心情决定', 3, 3),
(1, '在家追剧打游戏', 2, 4),
(1, '独处充电', 1, 5);

-- Seed options for question 2
INSERT INTO personality_options (question_id, option_text, score, sort_order) VALUES
(2, '全场焦点，带动气氛', 5, 1),
(2, '和熟悉的人聊得来', 4, 2),
(2, '观察为主，偶尔参与', 3, 3),
(2, '找安静角落待着', 2, 4),
(2, '能不去就不去', 1, 5);

-- Options for questions 3-10
INSERT INTO personality_options (question_id, option_text, score, sort_order) VALUES
(3, '耐心倾听，给出建议', 5, 1), (3, '会听但不太会安慰', 3, 2), (3, '不太擅长处理这种情况', 1, 3),
(4, '和谐氛围', 5, 1), (4, '效率第一', 3, 2), (4, '结果最重要', 1, 3),
(5, '提前完成', 5, 1), (5, '刚好赶上', 3, 2), (5, '经常需要延期', 1, 3),
(6, '井井有条', 5, 1), (6, '有点乱但能找到东西', 3, 2), (6, '完全混乱', 1, 3),
(7, '冷静分析解决', 1, 1), (7, '有点慌但能应对', 3, 2), (7, '容易焦虑', 5, 3),
(8, '不太在意', 1, 1), (8, '会反思一下', 3, 2), (8, '久久不能释怀', 5, 3),
(9, '非常兴奋想尝试', 5, 1), (9, '观望一下再说', 3, 2), (9, '更喜欢熟悉的事物', 1, 3),
(10, '探索未知的地方', 5, 1), (10, '有规划的自由行', 3, 2), (10, '去熟悉的地方更安心', 1, 3);

-- Seed avatar components
INSERT INTO avatar_components (category, name, image_url, rarity) VALUES
('face', '默认脸型A', '/assets/avatars/face_a.png', 1),
('face', '默认脸型B', '/assets/avatars/face_b.png', 1),
('face', '精致脸型', '/assets/avatars/face_premium.png', 2),
('hair', '短发A', '/assets/avatars/hair_short_a.png', 1),
('hair', '长发A', '/assets/avatars/hair_long_a.png', 1),
('hair', '潮流发色', '/assets/avatars/hair_trendy.png', 2),
('eyes', '圆眼', '/assets/avatars/eyes_round.png', 1),
('eyes', '凤眼', '/assets/avatars/eyes_phoenix.png', 1),
('eyes', '异色瞳', '/assets/avatars/eyes_hetero.png', 3),
('clothes', '卫衣', '/assets/avatars/clothes_hoodie.png', 1),
('clothes', '街头风', '/assets/avatars/clothes_street.png', 2),
('accessory', '耳机', '/assets/avatars/acc_headphone.png', 1),
('accessory', '墨镜', '/assets/avatars/acc_sunglasses.png', 1),
('accessory', '发光项链', '/assets/avatars/acc_neon.png', 3),
('background', '纯色背景', '/assets/avatars/bg_solid.png', 1),
('background', '渐变背景', '/assets/avatars/bg_gradient.png', 2);
