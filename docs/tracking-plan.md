# Spark 埋点事件文档 v1.0

> Phase 0 W1 交付 | 神策数据 SDK | 2026-06-17

---

## 1. 埋点架构

```
SDK: 神策数据（Sensors Analytics）免费版（100 万事件/月）
集成: Phase 1 W2 Flutter 端集成 sensors_analytics_flutter
设计原则:
  - 所有事件带公共属性（user_id、platform、app_version、timestamp）
  - 核心漏斗事件必须覆盖，非核心可后续补
  - 事件命名: snake_case，语义清晰
```

## 2. 公共属性（所有事件自动携带）

| 属性 | 类型 | 示例 | 说明 |
|---|---|---|---|
| `user_id` | string | "d7e743a5-..." | 注册后为 UUID，未注册为设备 ID |
| `is_guest` | bool | true/false | 是否未注册游客 |
| `platform` | string | "iOS"/"Android" | 操作系统 |
| `app_version` | string | "1.0.0" | App 版本号 |
| `city` | string | "上海" | 用户选择的城市 |
| `session_id` | string | UUID | 会话 ID，App 切后台 30min 后刷新 |

## 3. 核心漏斗事件

### 3.1 注册漏斗

| 事件名 | 触发时机 | 关键属性 |
|---|---|---|
| `app_open` | App 冷启动 | `source`（organic/search/referral） |
| `preview_card_view` | 预览卡片展示（未注册） | `card_index`（1-5）, `candidate_type`（bot/real） |
| `register_intercept_show` | 注册拦截弹窗展示 | `trigger`（swipe_right/send_message/view_zodiac/view_personality） |
| `register_start` | 点击注册按钮 | `method`（wechat/phone） |
| `register_complete` | 注册成功 | `method`, `duration_ms` |
| `real_name_verify_start` | 开始实名认证 | — |
| `real_name_verify_complete` | 实名认证通过 | `duration_ms` |
| `interest_tag_select` | 选择兴趣标签 | `tag_id`, `tag_name`, `category`, `selected_count` |
| `interest_tag_submit` | 提交兴趣标签 | `tag_ids`, `total_count` |
| `personality_quiz_start` | 开始人格测试 | — |
| `personality_quiz_answer` | 回答单题 | `question_id`, `option_id`, `question_index` |
| `personality_quiz_complete` | 完成人格测试 | `dimensions`（5 维度得分）, `duration_ms` |
| `personality_quiz_skip` | 跳过人格测试 | `current_question_index` |
| `avatar_select` | 选择虚拟形象 | `avatar_style_id` |
| `avatar_skip` | 跳过形象选择 | — |
| `onboarding_complete` | 引导流程全部完成 | `total_duration_ms`, `steps_completed` |

### 3.2 匹配漏斗

| 事件名 | 触发时机 | 关键属性 |
|---|---|---|
| `candidates_show` | 候选池卡片展示 | `candidate_count`, `bot_ratio` |
| `card_swipe_right` | 右滑喜欢 | `target_user_id`, `card_index`, `target_is_bot` |
| `card_swipe_left` | 左滑跳过 | `target_user_id`, `card_index`, `target_is_bot` |
| `filter_use` | 使用筛选 | `filter_type`（zodiac/interest/city）, `filter_value` |
| `match_new` | 双向匹配成功 | `match_id`, `target_user_id`, `compatibility_score`, `is_bot_match` |
| `match_popup_show` | 匹配弹窗展示 | `match_id`, `zodiac_compatibility` |
| `match_popup_chat` | 弹窗点击"打招呼" | `match_id` |
| `match_popup_continue` | 弹窗点击"继续刷" | `match_id` |
| `horoscope_card_view` | 运势卡片展示 | `zodiac_sign`, `daily_keyword` |
| `horoscope_cta_click` | 运势卡片点击"去看看推荐" | — |

### 3.3 聊天漏斗

| 事件名 | 触发时机 | 关键属性 |
|---|---|---|
| `chat_room_open` | 进入聊天室 | `room_id`, `source`（match_popup/chat_list/notification） |
| `icebreaker_show` | 破冰话题卡片展示 | `room_id`, `icebreaker_text`, `is_ai_generated` |
| `icebreaker_dismiss` | 关闭破冰话题 | `room_id` |
| `message_send` | 发送消息 | `room_id`, `client_msg_id`, `msg_type`（text/emoji）, `is_first_message` |
| `message_receive` | 收到消息 | `room_id`, `sender_id`, `msg_type` |
| `message_send_fail` | 消息发送失败 | `room_id`, `client_msg_id`, `error_reason` |
| `message_retry` | 点击重发 | `room_id`, `client_msg_id` |
| `chat_room_leave` | 离开聊天室 | `room_id`, `duration_ms`, `message_count_sent` |
| `user_block` | 拉黑用户 | `target_user_id`, `room_id`, `reason` |
| `user_report` | 举报用户 | `target_user_id`, `room_id`, `report_reason` |

### 3.4 留存与活跃

| 事件名 | 触发时机 | 关键属性 |
|---|---|---|
| `app_foreground` | App 进入前台 | — |
| `app_background` | App 进入后台 | `duration_ms`（本次会话时长） |
| `profile_view` | 查看"我的"页面 | — |
| `personality_card_share` | 分享人格卡片 | `share_channel`（wechat/moments/xiaohongshu/weibo） |
| `account_delete_request` | 申请注销账号 | — |
| `account_delete_confirm` | 确认注销 | — |

## 4. 非漏斗统计（用于 PM 周报）

| 事件名 | 说明 |
|---|---|
| `push_receive` | 收到推送通知 |
| `push_click` | 点击推送通知 |
| `ws_disconnect` | WebSocket 断开 |
| `ws_reconnect` | WebSocket 重连成功 |
| `ws_reconnect_fail` | WebSocket 重连失败 |
| `ai_service_fallback` | AI 服务降级到模板（`ai_service_enabled` 切换记录） |
| `api_error` | API 调用错误（含 endpoint、status_code） |

## 5. 看板设计

### 实时看板（分钟级）
- 当前在线用户数
- 过去 15 分钟消息发送量
- 过去 15 分钟新匹配数

### 日报看板
- DAU、新注册数、注册转化率
- 卡片展示数、右滑率、匹配率
- 新匹配对数（北极星指标）
- 人均发送消息数、破冰回复率
- 人格测试完成率、跳过率

### 漏斗看板
- 注册漏斗：app_open → register_intercept → register_complete → onboarding_complete
- 匹配漏斗：candidates_show → swipe_right → match_new → chat_room_open
- 聊天漏斗：chat_room_open → message_send → message_receive → message_send（reply）
- 留存漏斗：D1 → D3 → D7 → D30

### 预警阈值
- 次日留存 < 30%：企业微信告警
- 7 日留存 < 20%：邮件告警
- 匹配率 < 5%：企业微信告警
- WS 断连率 > 10%：企业微信告警
