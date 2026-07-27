package dao

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"agent/global"
	"agent/library/log"

	_ "github.com/go-sql-driver/mysql"
)

// MySQL MySQL连接管理
type MySQL struct {
	db *sql.DB
}

// NewMySQL 创建MySQL连接
func NewMySQL(cfg global.DatabaseConfig) (*MySQL, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接MySQL失败: %v", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpen)
	db.SetMaxIdleConns(cfg.MaxIdle)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("MySQL连接测试失败: %v", err)
	}

	return &MySQL{db: db}, nil
}

// Close 关闭连接
func (m *MySQL) Close() error {
	return m.db.Close()
}

// DB 获取底层数据库连接
func (m *MySQL) DB() *sql.DB {
	return m.db
}

// AutoMigrate 自动迁移数据库表
func (m *MySQL) AutoMigrate(ctx context.Context) error {
	// 创建用户表
	createUserTable := `
	CREATE TABLE IF NOT EXISTS users (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		username VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名',
		password VARCHAR(255) NOT NULL COMMENT '密码(bcrypt加密)',
		nickname VARCHAR(100) COMMENT '昵称',
		phone VARCHAR(20) UNIQUE NULL COMMENT '手机号',
		email VARCHAR(255) UNIQUE NULL COMMENT '邮箱',
		role VARCHAR(20) DEFAULT 'user' COMMENT '角色: admin/user',
		status TINYINT DEFAULT 1 COMMENT '状态: 1-正常 0-禁用',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_username (username),
		INDEX idx_phone (phone),
		INDEX idx_email (email)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';
	`

	_, err := m.db.ExecContext(ctx, createUserTable)
	if err != nil {
		log.Error(ctx, "AutoMigrate: 创建用户表失败 err=%v", err)
		return fmt.Errorf("创建用户表失败: %v", err)
	}

	createEmailCodeTable := `
	CREATE TABLE IF NOT EXISTS email_codes (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		email VARCHAR(255) NOT NULL COMMENT '邮箱',
		scene VARCHAR(32) NOT NULL COMMENT '场景: register/reset_password',
		code VARCHAR(20) NOT NULL COMMENT '验证码',
		expired_at DATETIME NOT NULL COMMENT '过期时间',
		used TINYINT DEFAULT 0 COMMENT '是否已使用',
		attempts INT DEFAULT 0 COMMENT '尝试次数',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_email_scene (email, scene),
		INDEX idx_expired_at (expired_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邮箱验证码表';
	`

	_, err = m.db.ExecContext(ctx, createEmailCodeTable)
	if err != nil {
		log.Error(ctx, "AutoMigrate: 创建邮箱验证码表失败 err=%v", err)
		return fmt.Errorf("创建邮箱验证码表失败: %v", err)
	}

	// 创建频率限制记录表
	createRateLimitTable := `
	CREATE TABLE IF NOT EXISTS rate_limits (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		user_id BIGINT NOT NULL COMMENT '用户ID',
		endpoint VARCHAR(50) NOT NULL COMMENT '接口标识',
		request_at DATETIME NOT NULL COMMENT '请求时间',
		INDEX idx_user_endpoint_time (user_id, endpoint, request_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='频率限制记录表';
	`

	_, err = m.db.ExecContext(ctx, createRateLimitTable)
	if err != nil {
		log.Error(ctx, "AutoMigrate: 创建频率限制表失败 err=%v", err)
		return fmt.Errorf("创建频率限制表失败: %v", err)
	}

	log.Info(ctx, "AutoMigrate: 数据库迁移完成")

	// 通用对话会话与消息表（深度思考模式使用）
	createChatSessionsTable := `
	CREATE TABLE IF NOT EXISTS chat_sessions (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		session_id VARCHAR(64) NOT NULL UNIQUE COMMENT '会话业务主键(UUID)',
		user_id BIGINT NOT NULL COMMENT '所属用户',
		title VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'AI 生成的会话标题',
		mode VARCHAR(16) NOT NULL DEFAULT 'normal' COMMENT '模式: normal/thinking/auto',
		last_message_at DATETIME NULL COMMENT '最后一条消息时间',
		status TINYINT NOT NULL DEFAULT 1 COMMENT '1-正常 0-归档 -1-软删',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_user_status (user_id, status, last_message_at DESC)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通用对话会话';
	`
	if _, err := m.db.ExecContext(ctx, createChatSessionsTable); err != nil {
		log.Error(ctx, "AutoMigrate: 创建 chat_sessions 失败 err=%v", err)
		return fmt.Errorf("创建 chat_sessions 失败: %v", err)
	}

	createChatMessagesTable := `
	CREATE TABLE IF NOT EXISTS chat_messages (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		session_id VARCHAR(64) NOT NULL COMMENT '关联的会话ID',
		role VARCHAR(16) NOT NULL COMMENT 'user/assistant/system',
		content MEDIUMTEXT NOT NULL COMMENT '消息内容',
		thought_content MEDIUMTEXT NULL COMMENT '思考过程（仅深度思考模式）',
		tool_calls TEXT NULL COMMENT '工具调用JSON',
		mode VARCHAR(16) NOT NULL DEFAULT 'normal' COMMENT '产生该消息的模式: normal/thinking/auto',
		status TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 1-成功 2-进行中 3-失败',
		token_count INT DEFAULT 0 COMMENT 'token 数量',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_session (session_id, id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通用对话消息';
	`
	if _, err := m.db.ExecContext(ctx, createChatMessagesTable); err != nil {
		log.Error(ctx, "AutoMigrate: 创建 chat_messages 失败 err=%v", err)
		return fmt.Errorf("创建 chat_messages 失败: %v", err)
	}
	log.Info(ctx, "AutoMigrate: 通用对话表(chat_sessions/chat_messages)迁移完成")

	// 兼容旧表：为已存在的 chat_messages 表补充 mode/status 字段
	addColumnIfNotExistsGeneric := func(tableName, columnName, columnDef string) {
		var count int
		err := m.db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?",
			tableName, columnName).Scan(&count)
		if err != nil {
			log.Warn(ctx, "AutoMigrate: 检查字段失败 table=%s, column=%s, err=%v", tableName, columnName, err)
			return
		}
		if count == 0 {
			_, err := m.db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tableName, columnDef))
			if err != nil {
				log.Warn(ctx, "AutoMigrate: 添加字段失败 table=%s, column=%s, err=%v", tableName, columnName, err)
			} else {
				log.Info(ctx, "AutoMigrate: 添加字段成功 table=%s, column=%s", tableName, columnName)
			}
		}
	}
	addColumnIfNotExistsGeneric("chat_messages", "mode",
		"mode VARCHAR(16) NOT NULL DEFAULT 'normal' COMMENT '产生该消息的模式: normal/thinking/auto'")
	addColumnIfNotExistsGeneric("chat_messages", "status",
		"status TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 1-成功 2-进行中 3-失败'")

	// 添加软删除字段（兼容已有表）
	addColumnIfNotExists := func(tableName, columnName, columnDef string) {
		var count int
		err := m.db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?",
			tableName, columnName).Scan(&count)
		if err != nil {
			log.Warn(ctx, "AutoMigrate: 检查字段失败 table=%s, column=%s, err=%v", tableName, columnName, err)
			return
		}
		if count == 0 {
			_, err := m.db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tableName, columnDef))
			if err != nil {
				log.Warn(ctx, "AutoMigrate: 添加字段失败 table=%s, column=%s, err=%v", tableName, columnName, err)
			} else {
				log.Info(ctx, "AutoMigrate: 添加字段成功 table=%s, column=%s", tableName, columnName)
			}
		}
	}

	addColumnIfNotExists("users", "is_deleted", "is_deleted TINYINT DEFAULT 0 COMMENT '是否删除：0-正常 1-删除'")
	addColumnIfNotExists("email_codes", "is_deleted", "is_deleted TINYINT DEFAULT 0 COMMENT '是否删除：0-正常 1-删除'")
	addColumnIfNotExists("rate_limits", "is_deleted", "is_deleted TINYINT DEFAULT 0 COMMENT '是否删除：0-正常 1-删除'")
	addColumnIfNotExists("sms_codes", "is_deleted", "is_deleted TINYINT DEFAULT 0 COMMENT '是否删除：0-正常 1-删除'")

	return nil
}

// AutoMigratePersonaTables 自动迁移角色卡相关表
func (m *MySQL) AutoMigratePersonaTables(ctx context.Context) error {
	// 创建角色卡表
	createPersonaTable := `
	CREATE TABLE IF NOT EXISTS personas (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		name VARCHAR(100) NOT NULL COMMENT '角色名称',
		avatar TEXT COMMENT '头像URL或base64',
		tagline VARCHAR(200) COMMENT '一句话介绍',
		personality TEXT COMMENT '性格特质描述',
		system_prompt TEXT NOT NULL COMMENT '系统提示词（核心人设）',
		creator_id BIGINT COMMENT '创建者用户ID，NULL表示系统预置',
		is_public BOOLEAN DEFAULT FALSE COMMENT '是否公开（其他用户可见）',
		usage_count INT DEFAULT 0 COMMENT '使用次数（用于热门排序）',
		status TINYINT DEFAULT 1 COMMENT '状态：1-正常 0-禁用 -1-删除',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_creator (creator_id),
		INDEX idx_public (is_public, status),
		INDEX idx_usage (usage_count DESC)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='AI角色卡';
	`

	_, err := m.db.ExecContext(ctx, createPersonaTable)
	if err != nil {
		log.Error(ctx, "AutoMigratePersonaTables: 创建角色卡表失败 err=%v", err)
		return fmt.Errorf("创建角色卡表失败: %v", err)
	}

	// 创建角色卡示例对话表
	createPersonaExampleTable := `
	CREATE TABLE IF NOT EXISTS persona_examples (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		persona_id BIGINT NOT NULL COMMENT '角色卡ID',
		user_input TEXT NOT NULL COMMENT '示例用户输入',
		ai_response TEXT NOT NULL COMMENT '示例AI回复',
		sort_order INT DEFAULT 0 COMMENT '排序序号',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_persona (persona_id, sort_order),
		FOREIGN KEY (persona_id) REFERENCES personas(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色卡示例对话';
	`

	_, err = m.db.ExecContext(ctx, createPersonaExampleTable)
	if err != nil {
		log.Error(ctx, "AutoMigratePersonaTables: 创建示例对话表失败 err=%v", err)
		return fmt.Errorf("创建示例对话表失败: %v", err)
	}

	// 创建角色卡对话历史表
	createPersonaChatTable := `
	CREATE TABLE IF NOT EXISTS persona_chats (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		user_id BIGINT NOT NULL COMMENT '用户ID',
		persona_id BIGINT NOT NULL COMMENT '角色卡ID',
		session_id VARCHAR(36) NOT NULL COMMENT '会话ID（UUID）',
		role VARCHAR(20) NOT NULL COMMENT '消息角色：user/assistant/system',
		content TEXT NOT NULL COMMENT '消息内容',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_user_session (user_id, session_id),
		INDEX idx_persona (persona_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色卡对话历史';
	`

	_, err = m.db.ExecContext(ctx, createPersonaChatTable)
	if err != nil {
		log.Error(ctx, "AutoMigratePersonaTables: 创建对话历史表失败 err=%v", err)
		return fmt.Errorf("创建对话历史表失败: %v", err)
	}

	log.Info(ctx, "AutoMigratePersonaTables: 角色卡相关表迁移完成")

	// 添加软删除字段（兼容已有表）
	addColumnIfNotExists := func(tableName, columnName, columnDef string) {
		var count int
		err := m.db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?",
			tableName, columnName).Scan(&count)
		if err != nil {
			log.Warn(ctx, "AutoMigratePersonaTables: 检查字段失败 table=%s, column=%s, err=%v", tableName, columnName, err)
			return
		}
		if count == 0 {
			_, err := m.db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tableName, columnDef))
			if err != nil {
				log.Warn(ctx, "AutoMigratePersonaTables: 添加字段失败 table=%s, column=%s, err=%v", tableName, columnName, err)
			} else {
				log.Info(ctx, "AutoMigratePersonaTables: 添加字段成功 table=%s, column=%s", tableName, columnName)
			}
		}
	}

	addColumnIfNotExists("persona_examples", "is_deleted", "is_deleted TINYINT DEFAULT 0 COMMENT '是否删除：0-正常 1-删除'")
	addColumnIfNotExists("persona_chats", "is_deleted", "is_deleted TINYINT DEFAULT 0 COMMENT '是否删除：0-正常 1-删除'")

	// 插入预置角色卡数据
	if err := m.insertPresetPersonas(ctx); err != nil {
		log.Warn(ctx, "AutoMigratePersonaTables: 插入预置角色卡失败 err=%v", err)
	}

	return nil
}

// insertPresetPersonas 插入预置角色卡数据
func (m *MySQL) insertPresetPersonas(ctx context.Context) error {
	// 检查是否已有数据
	var count int
	err := m.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM personas WHERE creator_id IS NULL").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		log.Info(ctx, "insertPresetPersonas: 已有预置角色卡，跳过插入")
		return nil
	}

	// 预置角色卡数据
	personas := []struct {
		name         string
		tagline      string
		personality  string
		systemPrompt string
		examples     []struct {
			userInput  string
			aiResponse string
		}
	}{
		{
			name:        "面试官 Lisa",
			tagline:     "严厉但专业的技术面试官",
			personality: "严肃、追问细节、给出建设性反馈",
			systemPrompt: `你是一位资深的技术面试官，有10年以上的面试经验。你的风格是：
1. 严肃但不刻薄，注重考察候选人的真实能力
2. 善于通过追问挖掘候选人的知识深度
3. 会指出回答中的不足，并给出改进建议
4. 关注基础知识和实际项目经验的结合
5. 面试结束时会给候选人一个总体评价和建议

你的面试流程通常是：自我介绍 → 项目经验 → 技术问题 → 算法/代码 → 反问环节`,
			examples: []struct {
				userInput  string
				aiResponse string
			}{
				{
					userInput:  "您好，我来参加面试",
					aiResponse: "你好，我是今天的面试官Lisa。先做个自我介绍吧，重点说说你的技术背景和最近的项目经历。",
				},
				{
					userInput:  "我熟悉MySQL数据库",
					aiResponse: "那我问你几个MySQL的问题。首先，你能说说MySQL的索引结构吗？B+树有什么特点？为什么MySQL选择B+树而不是B树？",
				},
			},
		},
		{
			name:        "写作教练",
			tagline:     "帮你提升写作能力的专业教练",
			personality: "温和、细致、注重逻辑",
			systemPrompt: `你是一位专业的写作教练，帮助用户提升写作能力。你的教学风格是：
1. 先理解用户的写作目的和目标读者
2. 从文章结构、逻辑、表达三个维度给出建议
3. 用具体的例子说明问题，而不是抽象的理论
4. 鼓励用户多写多练，给出具体的练习建议
5. 尊重用户的写作风格，不强行改变

你擅长：文章润色、结构优化、逻辑梳理、表达改进`,
			examples: []struct {
				userInput  string
				aiResponse string
			}{
				{
					userInput:  "帮我看看这篇文章怎么改",
					aiResponse: "好的，请把文章发给我。我会从结构、逻辑、表达三个维度帮你分析，并给出具体的修改建议。",
				},
				{
					userInput:  "我的文章总是写着写着就跑题了",
					aiResponse: "这个问题很常见，通常是开头的主题不够明确。建议你写之前先列个提纲，明确每一段的核心观点。可以试试「金字塔原理」——先说结论，再展开论证。",
				},
			},
		},
		{
			name:        "心理咨询师",
			tagline:     "温暖倾听，帮助你理清思绪",
			personality: "温暖、共情、引导思考",
			systemPrompt: `你是一位专业的心理咨询师，你的风格是：
1. 温暖包容，让来访者感到被理解
2. 不急于给建议，先帮助来访者理清思绪
3. 善于通过提问引导来访者自我探索
4. 保守来访者的隐私
5. 如遇严重心理问题，会建议寻求专业帮助

你遵循人本主义心理学理念，相信每个人都有自我成长的潜力。`,
			examples: []struct {
				userInput  string
				aiResponse string
			}{
				{
					userInput:  "最近感觉很焦虑，不知道该怎么办",
					aiResponse: "听起来你最近承受着一些压力。能跟我说说，让你感到焦虑的具体是什么事情吗？有时候把担忧说出来，就能让心里轻松一些。",
				},
				{
					userInput:  "我总是觉得自己不够好",
					aiResponse: "这种感觉一定很难受。我想问问你，你是从什么时候开始有这种想法的？当时发生了什么事情让你产生了这样的感受？",
				},
			},
		},
		{
			name:        "英语口语陪练",
			tagline:     "耐心纠正，鼓励你勇敢表达",
			personality: "耐心、鼓励、纠正错误",
			systemPrompt: `你是一位专业的英语口语陪练教练。你的教学风格是：
1. 营造轻松的对话氛围，让学习者敢于开口
2. 在对话中自然地纠正发音和语法错误
3. 根据学习者的水平调整语速和词汇
4. 鼓励学习者多说，提供扩展话题
5. 每次对话结束给出简短的反馈和建议

你的目标是帮助学习者建立说英语的自信心。`,
			examples: []struct {
				userInput  string
				aiResponse string
			}{
				{
					userInput:  "Hello, I want to practice English",
					aiResponse: "Hello! That's great that you want to practice. Don't worry about making mistakes - that's how we learn! What would you like to talk about today? Maybe we could start with something simple, like your hobbies or what you did yesterday?",
				},
				{
					userInput:  "I like reading book very much",
					aiResponse: "That's wonderful! Reading is a great hobby. By the way, I noticed you said \"reading book\" - in English, we usually say \"reading books\" (plural) or just \"reading\". So you could say \"I like reading books very much\" or \"I enjoy reading\". What kind of books do you like to read?",
				},
			},
		},
		{
			name:        "产品经理导师",
			tagline:     "培养你的产品思维和商业洞察力",
			personality: "专业、案例丰富、启发式提问",
			systemPrompt: `你是一位资深的产品经理导师，有丰富的互联网产品经验。你的教学风格是：
1. 用真实案例说明产品方法论
2. 善于通过提问引导思考，而不是直接给答案
3. 关注用户的真实需求，反对闭门造车
4. 强调数据驱动和用户反馈的重要性
5. 帮助学员建立系统的产品思维

你擅长：需求分析、用户研究、产品设计、数据分析、项目管理`,
			examples: []struct {
				userInput  string
				aiResponse string
			}{
				{
					userInput:  "我想做一个社交产品，能给点建议吗？",
					aiResponse: "做社交产品是一个很有挑战的方向。我想先问你几个问题：1. 你目标用户是谁？他们的社交需求是什么？2. 现有的社交产品为什么不能满足他们的需求？3. 你的产品有什么独特的价值主张？能先回答这些问题吗？",
				},
				{
					userInput:  "用户说他们想要这个功能",
					aiResponse: "用户说的不一定是他们真正需要的。有一个经典的故事：用户说他们想要更快的马，但福特给他们造了汽车。你有没有问过用户：他们遇到的问题是什么？为什么现有方案解决不了？真正的需求往往藏在表层之下。",
				},
			},
		},
	}

	// 插入角色卡
	for _, p := range personas {
		result, err := m.db.ExecContext(ctx,
			`INSERT INTO personas (name, tagline, personality, system_prompt, creator_id, is_public, usage_count, status) 
			 VALUES (?, ?, ?, ?, NULL, true, 0, 1)`,
			p.name, p.tagline, p.personality, p.systemPrompt)
		if err != nil {
			log.Error(ctx, "insertPresetPersonas: 插入角色卡失败 name=%s, err=%v", p.name, err)
			continue
		}

		personaID, _ := result.LastInsertId()

		// 插入示例对话
		for i, ex := range p.examples {
			_, err := m.db.ExecContext(ctx,
				`INSERT INTO persona_examples (persona_id, user_input, ai_response, sort_order) VALUES (?, ?, ?, ?)`,
				personaID, ex.userInput, ex.aiResponse, i)
			if err != nil {
				log.Error(ctx, "insertPresetPersonas: 插入示例失败 personaID=%d, err=%v", personaID, err)
			}
		}
	}

	log.Info(ctx, "insertPresetPersonas: 预置角色卡插入成功 count=%d", len(personas))
	return nil
}

// NbnhhshDAO 缩写词猜测数据访问
type NbnhhshDAO struct {
	mysql *MySQL
}

// NewNbnhhshDAO 创建NbnhhshDAO
func NewNbnhhshDAO(mysql *MySQL) *NbnhhshDAO {
	return &NbnhhshDAO{mysql: mysql}
}

// NbnhhshRecord 缩写词记录
type NbnhhshRecord struct {
	Name        string
	Trans       []string
	CreateTime  time.Time
	UpdatedTime time.Time
}

// GetByName 根据名称查询记录
func (d *NbnhhshDAO) GetByName(ctx context.Context, name string) (*NbnhhshRecord, error) {
	query := "SELECT name, trans, create_time, updated_time FROM nbnhhsh WHERE name = ?"

	var record NbnhhshRecord
	var transJSON string

	err := d.mysql.db.QueryRowContext(ctx, query, name).Scan(&record.Name, &transJSON, &record.CreateTime, &record.UpdatedTime)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		log.Error(ctx, "NbnhhshDAO.GetByName: 查询失败 name=%s, err=%v", name, err)
		return nil, err
	}

	if err := json.Unmarshal([]byte(transJSON), &record.Trans); err != nil {
		log.Error(ctx, "NbnhhshDAO.GetByName: 解析JSON失败 name=%s, err=%v", name, err)
		return nil, err
	}

	return &record, nil
}

// Save 保存记录
func (d *NbnhhshDAO) Save(ctx context.Context, name string, trans []string) error {
	transJSON, err := json.Marshal(trans)
	if err != nil {
		log.Error(ctx, "NbnhhshDAO.Save: 序列化JSON失败 name=%s, err=%v", name, err)
		return err
	}

	query := `
		INSERT INTO nbnhhsh (name, trans, create_time, updated_time) 
		VALUES (?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE trans = VALUES(trans), updated_time = NOW()
	`

	_, err = d.mysql.db.ExecContext(ctx, query, name, transJSON)
	if err != nil {
		log.Error(ctx, "NbnhhshDAO.Save: 保存失败 name=%s, err=%v", name, err)
		return err
	}

	log.Info(ctx, "NbnhhshDAO.Save: 保存成功 name=%s, trans=%v", name, trans)
	return nil
}

// IsCacheValid 检查缓存是否有效（3天内）
func (d *NbnhhshDAO) IsCacheValid(record *NbnhhshRecord) bool {
	if record == nil {
		return false
	}
	return time.Since(record.UpdatedTime) <= 3*24*time.Hour
}
