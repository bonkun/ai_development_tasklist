import React, { useState, useEffect } from 'react';
import { DragDropContext, Droppable, Draggable, DropResult } from '@hello-pangea/dnd';
import './App.css';
import { useNavigate } from 'react-router-dom';

type Task = {
  id: number;
  title: string;
  content: string;
  priority: '高' | '中' | '低';
  priority_name: string;
  due: string;
  progress_name: string;
  progress_id: number;
  position: number;
};

type Column = {
  name: string;
  items: Task[];
};

type Columns = {
  [key: string]: Column;
};

// 日付をYYYY-MM-DD形式に変換する関数を追加
const formatDateForInput = (dateString: string): string => {
  if (!dateString) return '';
  return dateString.split('T')[0];
};

const Kanban = () => {
  const navigate = useNavigate();

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) {
      alert('ログインしてください');
      navigate('/login');
    }
  }, [navigate]);

  const handleLogout = () => {
    localStorage.removeItem('token');
    navigate('/login');
  };

  const [columns, setColumns] = useState<Columns>({});
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editTask, setEditTask] = useState<Task | null>(null);
  const [newTask, setNewTask] = useState<Task>({
    id: 0,
    title: '',
    content: '',
    priority: '高',
    priority_name: '高',
    due: '',
    progress_name: '未着手',
    progress_id: 1,
    position: 0　//昇順でタスクが表示される。
  });

  useEffect(() => {
    const fetchTasks = async () => {
      try {
        const response = await fetch('http://localhost:8080/tasklist');
        if (!response.ok) {
          throw new Error('Failed to fetch tasks');
        }
        const data = await response.json();

        // Sort tasks by position in ascending order
        const sortedData = data.sort((a: Task, b: Task) => a.position - b.position);

        const columnsFromBackend = {
          '未着手': {
            name: '未着手',
            items: sortedData
              .filter((task: any) => task.progress_name === '未着手')
              .map((task: Task) => ({
                ...task,
                progress_id: getProgressIdByColumnId('未着手'),
                due: formatDateForInput(task.due)
              }))
          },
          '進行中': {
            name: '進行中',
            items: sortedData
              .filter((task: any) => task.progress_name === '進行中')
              .map((task: Task) => ({
                ...task,
                progress_id: getProgressIdByColumnId('進行中'),
                due: formatDateForInput(task.due)
              }))
          },
          '完了': {
            name: '完了',
            items: sortedData
              .filter((task: any) => task.progress_name === '完了')
              .map((task: Task) => ({
                ...task,
                progress_id: getProgressIdByColumnId('完了'),
                due: formatDateForInput(task.due)
              }))
          },
          '保留': {
            name: '保留',
            items: sortedData
              .filter((task: any) => task.progress_name === '保留')
              .map((task: Task) => ({
                ...task,
                progress_id: getProgressIdByColumnId('保留'),
                due: formatDateForInput(task.due)
              }))
          }
        };
        setColumns(columnsFromBackend);
      } catch (error) {
        console.error('Error fetching tasks:', error);
      }
    };

    fetchTasks();
  }, []);

  const onDragEnd = async (result: DropResult) => {
    if (!result.destination) return;
    const { source, destination } = result;
  
    const sourceColumn = columns[source.droppableId];
    const destColumn = columns[destination.droppableId];
  
    // 更新が必要なタスクを格納する配列
    let tasksToUpdate: Task[] = [];
  
    // 同じprogress内での移動の場合
    if (source.droppableId === destination.droppableId) {
      const copiedItems = [...sourceColumn.items];
      const [movedItem] = copiedItems.splice(source.index, 1);
      copiedItems.splice(destination.index, 0, movedItem);
  
      // 動後のリストで位置を振り直す
      const reorderedItems = copiedItems.map((task, index) => ({
        ...task,
        position: (index + 1) * 1000
      }));
  
      setColumns({
        ...columns,
        [source.droppableId]: {
          ...sourceColumn,
          items: reorderedItems
        }
      });
  
      // 同じカラム内の全タスクを更新対象とする
      tasksToUpdate = reorderedItems;
  
    } else {
      // 異なるprogress間での移動の場合
      const sourceItems = [...sourceColumn.items];
      const destItems = [...destColumn.items];
      const [movedItem] = sourceItems.splice(source.index, 1);
      
      // 移動するタスクのprogress情報を更新
      movedItem.progress_id = getProgressIdByColumnId(destination.droppableId);
      movedItem.progress_name = destination.droppableId;
      
      destItems.splice(destination.index, 0, movedItem);
  
      // 移動元と移動先の両方のリストで位置を振り直す
      const reorderedSourceItems = sourceItems.map((task, index) => ({
        ...task,
        position: (index + 1) * 1000
      }));
  
      const reorderedDestItems = destItems.map((task, index) => ({
        ...task,
        position: (index + 1) * 1000
      }));
  
      setColumns({
        ...columns,
        [source.droppableId]: {
          ...sourceColumn,
          items: reorderedSourceItems
        },
        [destination.droppableId]: {
          ...destColumn,
          items: reorderedDestItems
        }
      });
  
      // 移動元と移動先の両方のカラムのタスクを更新対象とする
      tasksToUpdate = [...reorderedSourceItems, ...reorderedDestItems];
    }
  
    // バックエンドへの一括更新処理
    try {
      console.log('Updating tasks:', tasksToUpdate);
      const response = await fetch('http://localhost:8080/update', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          tasks: tasksToUpdate.map(task => ({
            id: task.id,
            title: task.title,
            content: task.content,
            due: task.due,
            priority: task.priority,
            progress_id: task.progress_id,
            position: task.position
          }))
        })
      });
  
      if (!response.ok) {
        throw new Error('Failed to update tasks');
      }
  
      const result = await response.json();
      console.log('Update result:', result);
    } catch (error) {
      console.error('Error updating tasks:', error);
    }
  };


  const handleAddTask = async (progressId: number) => {
    try {
      const progressName = getProgressNameById(progressId);
      const columnItems = columns[progressName]?.items || [];
      
      // Calculate the new position
      const newPosition = columnItems.length === 0 
        ? 0 
        : columnItems[columnItems.length - 1].position + 1000;

      const response = await fetch('http://localhost:8080/insert', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          tasks: [{
            title: newTask.title,
            content: newTask.content,
            due: newTask.due,
            priority_name: newTask.priority,
            progress_id: progressId,
            position: newPosition
          }]
        })
      });

      if (!response.ok) {
        throw new Error('Failed to add task');
      }

      const result = await response.json();
      const newTaskId = result.successInserts[0].id;

      setColumns(prevColumns => ({
        ...prevColumns,
        [progressName]: {
          ...prevColumns[progressName],
          items: [
            ...prevColumns[progressName].items,
            { 
              ...newTask, 
              id: newTaskId, 
              due: formatDateForInput(newTask.due), 
              position: newPosition,
              progress_name: progressName
            }
          ]
        }
      }));

      setIsModalOpen(false);
      setNewTask({
        id: 0,
        title: '',
        content: '',
        priority: '高',
        priority_name: '高',
        due: '',
        progress_name: '未着手',
        progress_id: 1,
        position: 0
      });
    } catch (error) {
      console.error('Error adding task:', error);
    }
  };

  const handleTaskDoubleClick = (task: Task) => {
    console.log('Double-clicked task:', task);
    setEditTask({
      ...task,
      progress_id: getProgressIdByColumnId(task.progress_name),
      due: formatDateForInput(task.due),
      priority_name: task.priority_name
    });
    setIsModalOpen(true);
  };

  const handleUpdateTask = async () => {
    if (!editTask) return;
  
    console.log('Updating task:', editTask);
  
    try {
      const response = await fetch('http://localhost:8080/update', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          tasks: [{
            id: editTask.id,
            title: editTask.title,
            content: editTask.content,
            due: editTask.due,
            priority_name: editTask.priority_name,
            progress_id: editTask.progress_id,
            position: editTask.position
          }]
        })
      });
  
      if (!response.ok) {
        throw new Error('Failed to update task');
      }
  
      const result = await response.json();
      console.log('Update result:', result);
  
      setColumns(prevColumns => {
        const newColumns = { ...prevColumns };
        
        // まず、すべての列から該当タスクを削除
        Object.keys(newColumns).forEach(columnId => {
          newColumns[columnId] = {
            ...newColumns[columnId],
            items: newColumns[columnId].items.filter(item => item.id !== editTask.id)
          };
        });
  
        // 更新後のタスクを新しい列に追加
        const newProgressName = getProgressNameById(editTask.progress_id);
        const updatedTask = {
          ...editTask,
          progress_name: newProgressName,
          due: formatDateForInput(editTask.due),
          priority: editTask.priority_name as '高' | '中' | '低'
        };
  
        // 正しい位置に挿入
        const updatedItems = [...newColumns[newProgressName].items, updatedTask];
        updatedItems.sort((a, b) => a.position - b.position); // positionでソート
  
        newColumns[newProgressName] = {
          ...newColumns[newProgressName],
          items: updatedItems
        };
  
        return newColumns;
      });
  
      setIsModalOpen(false);
      setEditTask(null);
    } catch (error) {
      console.error('Error updating task:', error);
    }
  };

  // 追加: progress_idからprogress_nameを取得する関数
  const getProgressNameById = (id: number): string => {
    switch (id) {
      case 1:
        return '未着手';
      case 2:
        return '進行中';
      case 3:
        return '完了';
      case 4:
        return '保留';
      default:
        return '未着手';
    }
  };

  // 追加: columnIdかprogress_idを取得する関数
  const getProgressIdByColumnId = (columnId: string): number => {
    switch (columnId) {
      case '未着手':
        return 1;
      case '進行中':
        return 2;
      case '完了':
        return 3;
      case '保留':
        return 4;
      default:
        return 1; // デフォト値
    }
  };

  const calculateNewPosition = (items: Task[], index: number, movingTaskId: number): number => {
    console.log('=== Position Calculation Start ===');
    console.log('Items length:', items.length);
    console.log('Target index:', index);
    console.log('Moving task ID:', movingTaskId);

    const filteredItems = items.filter(item => item.id !== movingTaskId);

    console.log('=== Surrounding Tasks (after filtering) ===');
    const prevTask = index > 0 ? filteredItems[index - 1] : null;
    const nextTask = index < filteredItems.length ? filteredItems[index] : null;

    console.log('Previous task:', prevTask ? { id: prevTask.id, title: prevTask.title, position: prevTask.position } : 'none');
    console.log('Next task:', nextTask ? { id: nextTask.id, title: nextTask.title, position: nextTask.position } : 'none');

    if (filteredItems.length === 0) {
      return 5000;
    }

    if (index === 0) {
      return Number(filteredItems[0].position) - 5000;
    }

    if (index >= filteredItems.length) {
      return Number(filteredItems[filteredItems.length - 1].position) + 5000;
    }

    const beforePosition = prevTask?.position ?? 0;
    const afterPosition = nextTask?.position ?? 0;

    console.log('Calculating position between:', { beforePosition, afterPosition });

    return Math.floor(beforePosition + ((afterPosition - beforePosition) / 2));
  };


  return (
    <div className="App">
      <div className="kanban-container">
        <button onClick={handleLogout} className="logout-button">ログアウト</button>
        <h2>カンバンボード</h2>
      </div>
      <DragDropContext onDragEnd={onDragEnd}>
        {Object.entries(columns).map(([columnId, column], index) => {
          return (
            <div className="column" key={columnId}>
              <h3>{column.name}</h3>
              <Droppable droppableId={columnId} key={columnId}>
                {(provided, snapshot) => (
                  <div
                    {...provided.droppableProps}
                    ref={provided.innerRef}
                    className="droppable-col"
                  >
                    {column.items.map((task, index) => (
                      <Draggable key={task.id} draggableId={task.id.toString()} index={index}>
                        {(provided, snapshot) => (
                          <div
                            ref={provided.innerRef}
                            {...provided.draggableProps}
                            {...provided.dragHandleProps}
                            className="task-card"
                            onDoubleClick={() => handleTaskDoubleClick(task)}
                          >
                            <h3>{task.title}</h3>
                            <p>{task.content}</p>
                            <p>position: {task.position}</p>
                            <p>優先度: {task.priority_name}</p>
                            <p>期限: {task.due}</p>
                            <p>進捗: {task.progress_name}</p>
                          </div>
                        )}
                      </Draggable>
                    ))}
                    {provided.placeholder}
                    <button 
                      onClick={() => { 
                        setIsModalOpen(true); 
                        setNewTask({ ...newTask, progress_id: getProgressIdByColumnId(columnId) }); 
                      }} 
                      className="new-task-button"
                    >
                      +新規登録
                    </button>
                  </div>
                )}
              </Droppable>
            </div>
          );
        })}
      </DragDropContext>

      {isModalOpen && (
        <div className="modal">
          <h2>{editTask ? 'タスクを編集' : '新規タスク追加'}</h2>
          <input
            type="text"
            placeholder="タイトル"
            value={editTask ? editTask.title : newTask.title}
            onChange={(e) => {
              const value = e.target.value;
              if (editTask) {
                setEditTask({ ...editTask, title: value });
              } else {
                setNewTask({ ...newTask, title: value });
              }
            }}
          />
          <input
            type="text"
            placeholder="内容"
            value={editTask ? editTask.content : newTask.content}
            onChange={(e) => {
              const value = e.target.value;
              if (editTask) {
                setEditTask({ ...editTask, content: value });
              } else {
                setNewTask({ ...newTask, content: value });
              }
            }}
          />
          <select
            value={editTask ? editTask.priority_name : newTask.priority}
            onChange={(e) => {
              const value = e.target.value as '高' | '中' | '低';
              if (editTask) {
                setEditTask({ ...editTask, priority: value, priority_name: value });
              } else {
                setNewTask({ ...newTask, priority: value, priority_name: value });
              }
            }}
          >
            <option value="高">高</option>
            <option value="中">中</option>
            <option value="低">低</option>
          </select>
          <input
            type="date"
            value={editTask ? editTask.due : newTask.due}
            onChange={(e) => {
              const value = e.target.value;
              if (editTask) {
                setEditTask({ ...editTask, due: value });
              } else {
                setNewTask({ ...newTask, due: value });
              }
            }}
          />
          <button onClick={editTask ? handleUpdateTask : () => handleAddTask(newTask.progress_id)}>保存</button>
          <button onClick={() => { setIsModalOpen(false); setEditTask(null); }}>キャンセル</button>
        </div>
      )}
    </div>
  );
};

export default Kanban;
