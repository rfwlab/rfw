import { createSignal, createMemo, For } from 'solid-js';
import { createStore } from 'solid-js/store';

export default function App() {
  const [todos, setTodos] = createStore([]);
  const [filter, setFilter] = createSignal('all');
  const [title, setTitle] = createSignal('');
  const [live, setLive] = createSignal(0);
  let nextId = 0;

  const visible = createMemo(() =>
    todos.filter((t) =>
      filter() === 'active' ? !t.done : filter() === 'done' ? t.done : true
    )
  );
  const active = createMemo(() => todos.filter((t) => !t.done).length);

  function addTodo() {
    if (!title()) return;
    setTodos(todos.length, { id: ++nextId, title: title(), done: false });
    setTitle('');
  }

  function onKeydown(e) {
    if (e.key === 'Enter') addTodo();
  }

  function toggle(id) {
    setTodos((t) => t.id === id, 'done', (d) => !d);
  }

  function remove(id) {
    setTodos((prev) => prev.filter((t) => t.id !== id));
  }

  function startLive() {
    setLive(0);
    let i = 0;
    const timer = setInterval(() => {
      setLive(++i);
      if (i >= 30) clearInterval(timer);
    }, 100);
  }

  return (
    <div class="todoapp">
      <h1>todos</h1>
      <div class="entry">
        <input
          id="new-todo"
          type="text"
          placeholder="What needs to be done?"
          value={title()}
          onInput={(e) => setTitle(e.target.value)}
          onKeyDown={onKeydown}
        />
        <button id="add-todo" onClick={addTodo}>Add</button>
      </div>
      <ul id="todo-list">
        <For each={visible()}>
          {(t) => (
            <li class={t.done ? 'done' : ''} data-id={t.id}>
              <input type="checkbox" checked={t.done} onChange={() => toggle(t.id)} />
              <span>{t.title}</span>
              <button class="destroy" onClick={() => remove(t.id)}>x</button>
            </li>
          )}
        </For>
      </ul>
      <footer>
        <span id="todo-count">{active()} items left</span>
        <button onClick={() => setFilter('all')}>All</button>
        <button onClick={() => setFilter('active')}>Active</button>
        <button onClick={() => setFilter('done')}>Done</button>
      </footer>
      <div class="live">
        <button id="start-live" onClick={startLive}>Start live</button>
        <span id="live-counter">{live()}</span>
      </div>
    </div>
  );
}
