
# Thai message templates for all operation types (using frontend drag & drop style)
MESSAGES = {
    # Class definitions
    "class_defined": "คลาส {class_name} ถูกกำหนด",
    
    # Instance creation
    "instance_created_singly": "สร้าง singly linked list '{var_name}' (ว่างเปล่า)",
    "instance_created_doubly": "สร้าง doubly linked list '{var_name}' (ว่างเปล่า)",
    "instance_created_generic": "สร้าง {class_name} '{var_name}'",
    "node_created": "สร้างโหนด '{var_name}' ด้วยข้อมูล '{node_name}'",
    
    # Attribute operations
    "attribute_set_first": "ตั้งค่า {instance_name}.{attribute} เป็น {value_var} (โหนดแรก)",
    "attribute_set": "อัปเดต {instance_name}.{attribute} เป็น {value_var}",
    "attribute_set_generic": "ตั้งค่า {instance_name}.{attribute} เป็น {value_var}",
    "chained_attribute_set": "เชื่อมต่อ {instance_name}.head.next เป็น {value_var}",
    "prev_attribute_set": "เชื่อมต่อ {node_var}.prev เป็น {instance_name}.head",
    "chained_attribute_failed": "ไม่สามารถเชื่อมต่อ {instance_name}.head.next เป็น {value_var}",
    "prev_attribute_failed": "ไม่สามารถเชื่อมต่อ {node_var}.prev เป็น {instance_name}.head",
    
    # Insert operations (using frontend drag & drop style)
    "insert_first_node": "เพิ่มข้อมูล '{data}' ที่ตำแหน่งเริ่มต้นของ linked list",
    "insert_front": "เพิ่มข้อมูล '{data}' ที่ตำแหน่งเริ่มต้นของ linked list",
    "insert_last": "เพิ่มข้อมูล '{data}' ที่ตำแหน่งท้ายของ linked list",
    "insert_before": "เพิ่มข้อมูล '{new_data}' ก่อน '{target_name}' ใน linked list",
    "insert_before_head": "เพิ่มข้อมูล '{new_data}' ก่อน '{target_name}' ใน linked list",
    "target_not_found": "ไม่พบข้อมูล '{target_name}' ใน linked list",
    
    # Delete operations (using frontend drag & drop style)
    "delete_from_head": "ลบข้อมูลที่ตำแหน่งเริ่มต้นของ linked list",
    "delete_node": "ลบข้อมูล '{target_name}' ออกจาก linked list",
    "delete_empty_list": "ไม่สามารถลบข้อมูล: linked list ว่างเปล่า",
    "delete_target_not_found": "ไม่พบข้อมูล '{target_name}' ใน linked list",
    
    # Traverse operations (using frontend drag & drop style)
    "traverse_empty": "linked list ว่างเปล่า",
    "traverse_output": "-> {output}",
    "traverse_method": "เดินทางผ่าน linked list ทั้งหมด: {output}",
    "traverse_method_empty": "เดินทางผ่าน linked list ทั้งหมด: linked list ว่างเปล่า",
    
    # Method execution wrappers
    "using_method": "โดยใช้ {method_name}()",
    "insert_front_using": "เพิ่มข้อมูล '{data}' ที่ตำแหน่งเริ่มต้นของ linked list โดยใช้ {method_name}()",
    "insert_last_using": "เพิ่มข้อมูล '{data}' ที่ตำแหน่งท้ายของ linked list โดยใช้ {method_name}()",
    "insert_before_using": "เพิ่มข้อมูล '{new_data}' ก่อน '{target_name}' ใน linked list โดยใช้ {method_name}()",
    "delete_using": "ลบข้อมูล '{target_name}' ออกจาก linked list โดยใช้ {method_name}()",
    
    # Error messages
    "instance_not_found": "ไม่พบ linked list '{instance_name}'",
    "variable_not_found": "ไม่พบตัวแปร '{var_name}'",
    "nodes_corrupted": "ข้อมูลโหนดเสียหาย",
    "invalid_instance": "linked list ไม่ถูกต้อง",
    "method_not_found": "ไม่พบเมธอด {method_name} ใน {class_type}",
    "method_requires_data": "{method_name} ต้องการข้อมูล",
    "method_requires_target": "{method_name} ต้องการข้อมูลเป้าหมาย",
    "method_requires_params": "{method_name} ต้องการพารามิเตอร์",
    "constructor_called": "เรียกคอนสตรักเตอร์ {method_name} สำหรับ {instance_name}",
    "getter_returned": "{instance_name}.{method_name}(): คืนค่า",
    "executed_unknown": "ดำเนินการ {method_name} บน {instance_name} (พฤติกรรม: {behavior_type})",
    
    # Runtime error messages with line numbers
    "syntax_error": "เกิดข้อผิดพลาดทางไวยากรณ์ที่บรรทัด {line}: {message}",
    "name_error": "ไม่พบตัวแปร '{name}' ที่บรรทัด {line}",
    "attribute_error": "ไม่มี attribute '{attr}' ใน '{obj}' ที่บรรทัด {line}",
    "type_error": "ประเภทข้อมูลไม่ถูกต้องที่บรรทัด {line}: {message}",
    "value_error": "ค่าข้อมูลไม่ถูกต้องที่บรรทัด {line}: {message}",
    "index_error": "index เกินขอบเขตที่บรรทัด {line}",
    "runtime_error": "เกิดข้อผิดพลาดขณะรันโค้ดที่บรรทัด {line}: {message}",
    
    # Stack operations (using frontend drag & drop style)
    # Stack operations (Educational & Detailed)
    "stack_push": "นำข้อมูล '{data}' วางซ้อนทับลงไปบน Stack ข้อมูลใหม่นี้จะกลายเป็น **Top** (ข้อมูลบนสุด) ทันที ส่วนข้อมูลเดิมจะถูกดันลงไปด้านล่าง ส่งผลให้จำนวนข้อมูล (Size) เพิ่มขึ้นเป็น {size} ตามหลักการ **LIFO (Last In, First Out)** หรือ 'เข้าทีหลัง ออกก่อน'",
    "stack_pop": "ดึงข้อมูล '{data}' ที่อยู่ตำแหน่งบนสุด (Top) ออกจาก Stack ซึ่งข้อมูลนี้คือตัวล่าสุดที่ถูกใส่เข้ามา เมื่อนำออกไปแล้ว ข้อมูลที่อยู่ถัดลงไปจะกลายเป็น Top ตัวใหม่แทน จำนวนข้อมูลลดลงเหลือ {size}",
    "stack_peek": "ดูข้อมูลที่ตำแหน่งบนสุด (Top) ของ Stack โดยไม่มีการนำข้อมูลออก เพื่อตรวจสอบค่าล่าสุดที่ถูกเพิ่มเข้ามา ค่าปัจจุบันคือ '{data}'",
    "stack_is_empty": "ตรวจสอบสถานะของ Stack ว่าว่างเปล่าไม่มีข้อมูลอยู่หรือไม่ ผลลัพธ์คือ `{result}`",
    "stack_size": "นับจำนวนข้อมูลทั้งหมดที่เก็บอยู่ใน Stack ปัจจุบัน ได้ผลลัพธ์คือ {size} ตัว",
    
    # Binary Search Tree operations (using frontend drag & drop style)
    "bst_insert": "เพิ่มข้อมูล '{data}' เข้าไปใน binary search tree",
    "bst_delete": "ลบข้อมูล '{data}' ออกจาก binary search tree",
    "bst_search": "ค้นหาข้อมูล '{data}' ใน binary search tree",
    "bst_traverse_inorder": "เดินทางผ่าน binary search tree แบบ inorder",
    "bst_traverse_preorder": "เดินทางผ่าน binary search tree แบบ preorder",
    "bst_traverse_postorder": "เดินทางผ่าน binary search tree แบบ postorder",
    
    # Graph operations (using frontend drag & drop style)
    "graph_add_vertex": "เพิ่ม vertex '{data}' เข้าไปในกราฟ",
    "graph_add_edge": "เพิ่ม edge จาก '{from_vertex}' ไป '{to_vertex}'",
    "graph_remove_vertex": "ลบ vertex '{data}' และ edge ที่เชื่อมกับมัน",
    "graph_remove_edge": "ลบ edge จาก '{from_vertex}' ไป '{to_vertex}'",
    "graph_traversal_dfs": "เดินทางผ่านกราฟด้วย DFS เริ่มจาก '{start_vertex}'",
    "graph_traversal_bfs": "เดินทางผ่านกราฟด้วย BFS เริ่มจาก '{start_vertex}'",
    "graph_shortest_path": "หาเส้นทางที่สั้นที่สุดจาก '{start_vertex}' ไป '{end_vertex}'",
    
    # Legacy operations
    "linkedlist_initialized": "เริ่มต้น linked list ว่างเปล่า",
    "executed_line": "ดำเนินการ: {line}",
}


def get_message(key: str, **kwargs) -> str:
    """
    Get a Thai message by key with optional formatting parameters.
    
    Args:
        key: Message key from MESSAGES dictionary
        **kwargs: Formatting parameters for the message template
        
    Returns:
        Formatted Thai message string
        
    Raises:
        KeyError: If message key is not found
    """
    if key not in MESSAGES:
        raise KeyError(f"Message key '{key}' not found in MESSAGES dictionary")
    
    return MESSAGES[key].format(**kwargs)


def get_class_defined_message(class_name: str) -> str:
    """Get class definition message in Thai."""
    return get_message("class_defined", class_name=class_name)


def get_instance_created_message(class_name: str, var_name: str, is_doubly: bool = False) -> str:
    """Get instance creation message in Thai."""
    if is_doubly:
        return get_message("instance_created_doubly", class_name=class_name, var_name=var_name)
    else:
        return get_message("instance_created_singly", class_name=class_name, var_name=var_name)


def get_node_created_message(var_name: str, node_name: str) -> str:
    """Get node creation message in Thai."""
    return get_message("node_created", var_name=var_name, node_name=node_name)


def get_insert_message(operation_type: str, data: str, count: int = None, 
                      target_name: str = None, method_name: str = None) -> str:
    """Get insert operation message in Thai."""
    if operation_type == "first":
        return get_message("insert_first_node", data=data)
    elif operation_type == "front":
        if method_name:
            return get_message("insert_front_using", data=data, count=count, method_name=method_name)
        else:
            return get_message("insert_front", data=data, count=count)
    elif operation_type == "last":
        if method_name:
            return get_message("insert_last_using", data=data, count=count, method_name=method_name)
        else:
            return get_message("insert_last", data=data, count=count)
    elif operation_type == "before":
        if method_name:
            return get_message("insert_before_using", new_data=data, target_name=target_name, 
                             count=count, method_name=method_name)
        else:
            return get_message("insert_before", new_data=data, target_name=target_name, count=count)
    else:
        return get_message("insert_front", data=data, count=count)


def get_delete_message(target_name: str, count: int, from_head: bool = False, 
                      method_name: str = None) -> str:
    """Get delete operation message in Thai."""
    if method_name:
        return get_message("delete_using", target_name=target_name, count=count, method_name=method_name)
    elif from_head:
        return get_message("delete_from_head", target_name=target_name, count=count)
    else:
        return get_message("delete_node", target_name=target_name, count=count)


def get_traverse_message(instance_name: str, method_name: str, output: str = None) -> str:
    """Get traverse operation message in Thai."""
    if output is None or output == "":
        return get_message("traverse_method_empty", instance_name=instance_name, method_name=method_name)
    else:
        return get_message("traverse_method", instance_name=instance_name, method_name=method_name, output=output)


def get_error_message(error_type: str, **kwargs) -> str:
    """Get error message in Thai."""
    return get_message(error_type, **kwargs)


def get_stack_message(operation_type: str, data: str = None, **kwargs) -> str:
    """Get stack operation message in Thai."""
    if operation_type == "push":
        return get_message("stack_push", data=data)
    elif operation_type == "pop":
        return get_message("stack_pop")
    elif operation_type == "peek":
        return get_message("stack_peek")
    elif operation_type == "is_empty":
        return get_message("stack_is_empty")
    elif operation_type == "size":
        return get_message("stack_size")
    else:
        return get_message("executed_unknown", method_name=operation_type, instance_name="stack", behavior_type="stack")


def get_bst_message(operation_type: str, data: str = None, **kwargs) -> str:
    """Get BST operation message in Thai."""
    if operation_type == "insert":
        return get_message("bst_insert", data=data)
    elif operation_type == "delete":
        return get_message("bst_delete", data=data)
    elif operation_type == "search":
        return get_message("bst_search", data=data)
    elif operation_type == "traverse_inorder":
        return get_message("bst_traverse_inorder")
    elif operation_type == "traverse_preorder":
        return get_message("bst_traverse_preorder")
    elif operation_type == "traverse_postorder":
        return get_message("bst_traverse_postorder")
    else:
        return get_message("executed_unknown", method_name=operation_type, instance_name="bst", behavior_type="bst")


def get_graph_message(operation_type: str, data: str = None, from_vertex: str = None, 
                     to_vertex: str = None, start_vertex: str = None, end_vertex: str = None, **kwargs) -> str:
    """Get graph operation message in Thai."""
    if operation_type == "add_vertex":
        return get_message("graph_add_vertex", data=data)
    elif operation_type == "add_edge":
        return get_message("graph_add_edge", from_vertex=from_vertex, to_vertex=to_vertex)
    elif operation_type == "remove_vertex":
        return get_message("graph_remove_vertex", data=data)
    elif operation_type == "remove_edge":
        return get_message("graph_remove_edge", from_vertex=from_vertex, to_vertex=to_vertex)
    elif operation_type == "traversal_dfs":
        return get_message("graph_traversal_dfs", start_vertex=start_vertex)
    elif operation_type == "traversal_bfs":
        return get_message("graph_traversal_bfs", start_vertex=start_vertex)
    elif operation_type == "shortest_path":
        return get_message("graph_shortest_path", start_vertex=start_vertex, end_vertex=end_vertex)
    else:
        return get_message("executed_unknown", method_name=operation_type, instance_name="graph", behavior_type="graph")
