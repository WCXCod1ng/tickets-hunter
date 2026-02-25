import time
from datetime import datetime, timedelta

def generate_sql():
    # 1. 生成每次运行都不一样的唯一 ID (基于毫秒级时间戳)
    # 这样可以保证你多次运行生成的测试用例不会互相冲突
    event_id = int(time.time() * 1000)

    # 2. 动态计算时间 (保证任何时候运行，状态都是"正在售卖"且"还未演出")
    now = datetime.now()
    sale_start_time = now - timedelta(days=1)   # 昨天开售
    sale_end_time = now + timedelta(days=30)    # 30天后停售
    show_time = now + timedelta(days=35)        # 35天后演出

    time_format = "%Y-%m-%d %H:%M:%S"

    # --- 构造 Insert 语句 ---
    insert_lines = []
    insert_lines.append(f"USE `ticket_master`;\n")
    insert_lines.append("-- ==========================================")
    insert_lines.append(f"-- 自动生成测试数据, Event ID: {event_id}")
    insert_lines.append(f"-- 生成时间: {now.strftime(time_format)}")
    insert_lines.append("-- ==========================================\n")

    # 插入场次表
    insert_lines.append("-- 1. 插入演唱会场次")
    event_sql = f"""INSERT INTO `ticket_event`
(`id`, `title`, `cover_url`, `show_time`, `venue`, `sale_start_time`, `sale_end_time`, `status`)
VALUES
({event_id}, 'Go-Zero 抢票架构演示演唱会-北京站', 'https://dummyimage.com/600x800', '{show_time.strftime(time_format)}', '鸟巢国家体育场', '{sale_start_time.strftime(time_format)}', '{sale_end_time.strftime(time_format)}', 1);
"""
    insert_lines.append(event_sql)

    # 插入座位表
    insert_lines.append("-- 2. 插入座位表 (3个区域，每个区域 3行 x 5列)")
    insert_lines.append("INSERT INTO `ticket_seat` (`event_id`, `seat_type`, `section`, `row_no`, `seat_no`, `price`, `status`, `version`) VALUES ")

    # 定义分区配置: 名称, 座位类型(1普通/2VIP/3内场), 价格
    sections = [
        {"name": "A区", "type": 3, "price": 128000},
        {"name": "B区", "type": 2, "price": 88000},
        {"name": "C区", "type": 1, "price": 48000}
    ]

    seat_values = []
    for sec in sections:
        for row in range(1, 4):     # 1 到 3 排
            for col in range(1, 6): # 1 到 5 号
                seat_values.append(f"({event_id}, {sec['type']}, '{sec['name']}', {row}, {col}, {sec['price']:.2f}, 0, 0)")

    # 拼接所有座位 values，最后一个以分号结尾
    insert_lines.append(",\n".join(seat_values) + ";\n")

    # --- 构造 Delete 语句 ---
    delete_lines = []
    delete_lines.append(f"USE `ticket_master`;\n")
    delete_lines.append("-- ==========================================")
    delete_lines.append(f"-- 清理测试数据, Event ID: {event_id}")
    delete_lines.append("-- ==========================================\n")
    # 注意删除顺序：先删子表（座位），再删主表（场次）
    delete_lines.append(f"DELETE FROM `ticket_seat` WHERE `event_id` = {event_id};")
    delete_lines.append(f"DELETE FROM `ticket_event` WHERE `id` = {event_id};")
    delete_lines.append("SELECT '清理完成' AS result;")

    # --- 写入文件 ---
    insert_filename = "insert_tickets.sql"
    delete_filename = "delete_tickets.sql"

    with open(insert_filename, "w", encoding="utf-8") as f:
        f.write("\n".join(insert_lines))

    with open(delete_filename, "w", encoding="utf-8") as f:
        f.write("\n".join(delete_lines))

    print("✅ 脚本执行成功！")
    print(f"👉 生成的 Event ID: {event_id}")
    print(f"👉 已生成数据插入脚本: {insert_filename} (包含 1个场次，45 个座位)")
    print(f"👉 已生成数据清理脚本: {delete_filename}")

if __name__ == "__main__":
    generate_sql()