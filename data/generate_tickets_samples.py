import time
from datetime import datetime, timedelta

def generate_sql():
    # ===================================================================
    #                           可配置区域
    # ===================================================================
    # 1. 分区配置: (分区名, 座位类型, 价格(单位:分))
    #    - 座位类型: 1=普通, 2=VIP, 3=内场
    #    - 你可以按需增加或修改分区
    SECTIONS_CONFIG = [
        ('A', 3, 188800), # 内场
        ('B', 3, 158800), # 内场
        ('C', 2, 128800), # 看台VIP
        ('D', 2, 98800),  # 看台VIP
        ('E', 1, 68800),  # 普通看台
        ('F', 1, 48800),  # 普通看台
    ]

    # 2. 每个分区的座位数量
    ROWS_PER_SECTION = 20  # 每区 20 排
    SEATS_PER_ROW = 50   # 每排 50 座
    # ===================================================================


    # 1. 生成每次运行都不一样的唯一 ID (基于毫秒级时间戳)
    event_id = int(time.time() * 1000)

    # 2. 动态计算时间 (保证任何时候运行，状态都是"正在售卖"且"还未演出")
    now = datetime.now()
    sale_start_time = now - timedelta(days=1)   # 昨天开售
    sale_end_time = now + timedelta(days=30)    # 30天后停售
    show_time = now + timedelta(days=35)        # 35天后演出

    time_format = "%Y-%m-%d %H:%M:%S"

    # --- 构造 Insert 语句 ---
    insert_lines = []
    insert_lines.append(f"USE `tickets_hunter`;\n")
    insert_lines.append("-- ==========================================")
    insert_lines.append(f"-- 自动生成测试数据, Event ID: {event_id}")
    insert_lines.append(f"-- 生成时间: {now.strftime(time_format)}")
    insert_lines.append("-- ==========================================\n")

    # 插入场次表
    insert_lines.append("-- 1. 插入演唱会场次")
    event_sql = f"""INSERT INTO `ticket_event`
(`id`, `title`, `cover_url`, `show_time`, `venue`, `sale_start_time`, `sale_end_time`, `status`)
VALUES
({event_id}, 'Go-Zero 抢票架构演示演唱会-体育场站', 'https://dummyimage.com/600x800', '{show_time.strftime(time_format)}', '鸟巢国家体育场', '{sale_start_time.strftime(time_format)}', '{sale_end_time.strftime(time_format)}', 1);
"""
    insert_lines.append(event_sql)

    # 插入座位表
    total_seats = 0
    num_sections = len(SECTIONS_CONFIG)
    insert_lines.append(f"\n-- 2. 插入座位表 ({num_sections}个分区, 每个分区 {ROWS_PER_SECTION}行 x {SEATS_PER_ROW}列)")
    insert_lines.append("INSERT INTO `ticket_seat` (`event_id`, `seat_type`, `section`, `seat_index`, `row_no`, `seat_no`, `price`, `status`, `version`) VALUES ")

    seat_values = []
    for sec_name, sec_type, sec_price in SECTIONS_CONFIG:
        # 每个分区内的 seat_index 都从 0 开始计数
        seat_index_counter = 0
        for row in range(1, ROWS_PER_SECTION + 1):
            for col in range(1, SEATS_PER_ROW + 1):
                # 构造 VALUES 子句，注意增加了 seat_index_counter
                seat_values.append(f"({event_id}, {sec_type}, '{sec_name}', {seat_index_counter}, {row}, {col}, {sec_price}, 0, 0)")
                seat_index_counter += 1
                total_seats += 1

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
    insert_filename = f"insert_tickets_{event_id}.sql"
    delete_filename = f"delete_tickets_{event_id}.sql"

    with open(insert_filename, "w", encoding="utf-8") as f:
        f.write("\n".join(insert_lines))

    with open(delete_filename, "w", encoding="utf-8") as f:
        f.write("\n".join(delete_lines))

    print("✅ 脚本执行成功！")
    print(f"👉 生成的 Event ID: {event_id}")
    print(f"👉 已生成数据插入脚本: {insert_filename} (包含 1个场次, {num_sections}个分区, 共 {total_seats} 个座位)")
    print(f"👉 已生成数据清理脚本: {delete_filename}")

if __name__ == "__main__":
    generate_sql()