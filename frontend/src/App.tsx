import React, { useState, useEffect } from 'react';
import { DragDropContext, Droppable, Draggable, DropResult } from '@hello-pangea/dnd';
import './App.css';

type Task = {
  id: number;
  title: string;
  content: string;
  priority: '高' | '中' | '低';
  due: string;
  progress_name: string;
};

type Column = {
  name: string;
  items: Task[];
};

type Columns = {
  [key: string]: Column;
};

const App = () => {
  const [columns, setColumns] = useState<Columns>({});

  useEffect(() => {
    const fetchTasks = async () => {
      try {
        const response = await fetch('http://localhost:8080/tasklist');
        if (!response.ok) {
          throw new Error('Failed to fetch tasks');
        }
        const data = await response.json();
        const columnsFromBackend = {
          '未着手': {
            name: '未着手',
            items: data.filter((task: Task) => task.progress_name === '未着手')
          },
          '進行中': {
            name: '進行中',
            items: data.filter((task: Task) => task.progress_name === '進行中')
          },
          '完了': {
            name: '完了',
            items: data.filter((task: Task) => task.progress_name === '完了')
          },
          '保留': {
            name: '保留',
            items: data.filter((task: Task) => task.progress_name === '保留')
          }
        };
        setColumns(columnsFromBackend);
      } catch (error) {
        console.error('Error fetching tasks:', error);
      }
    };

    fetchTasks();
  }, []);

  const onDragEnd = (result: DropResult) => {
    if (!result.destination) return;
    const { source, destination } = result;

    const sourceColumn = columns[source.droppableId as keyof typeof columns];
    const destColumn = columns[destination.droppableId as keyof typeof columns];

    if (sourceColumn === destColumn) {
      // 同じカラム内での移動
      const newItems = Array.from(sourceColumn.items);
      const [movedItem] = newItems.splice(source.index, 1);
      newItems.splice(destination.index, 0, movedItem);

      setColumns({
        ...columns,
        [source.droppableId]: {
          ...sourceColumn,
          items: newItems
        }
      });
    } else {
      // 異なるカラム間での移動
      const sourceItems = Array.from(sourceColumn.items);
      const destItems = Array.from(destColumn.items);
      const [movedItem] = sourceItems.splice(source.index, 1);

      destItems.splice(destination.index, 0, movedItem);

      setColumns({
        ...columns,
        [source.droppableId]: {
          ...sourceColumn,
          items: sourceItems
        },
        [destination.droppableId]: {
          ...destColumn,
          items: destItems
        }
      });
    }
  };

  return (
    <div className="App">
      <h2>Task List</h2>
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
                          >
                            <h3>{task.title}</h3>
                            <p>{task.content}</p>
                            <p>優先度: {task.priority}</p>
                            <p>期限: {task.due}</p>
                            <p>進捗: {task.progress_name}</p>
                          </div>
                        )}
                      </Draggable>
                    ))}
                    {provided.placeholder}
                  </div>
                )}
              </Droppable>
            </div>
          );
        })}
      </DragDropContext>
    </div>
  );
};

export default App;
