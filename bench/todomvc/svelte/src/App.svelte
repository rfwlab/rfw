<script>
  let todos = $state([]);
  let filter = $state('all');
  let title = $state('');
  let live = $state(0);
  let nextId = 0;

  const visible = $derived(
    todos.filter((t) =>
      filter === 'active' ? !t.done : filter === 'done' ? t.done : true
    )
  );
  const active = $derived(todos.filter((t) => !t.done).length);

  function addTodo() {
    if (!title) return;
    todos.push({ id: ++nextId, title, done: false });
    title = '';
  }

  function onKeydown(e) {
    if (e.key === 'Enter') addTodo();
  }

  function toggle(id) {
    const t = todos.find((t) => t.id === id);
    if (t) t.done = !t.done;
  }

  function remove(id) {
    const i = todos.findIndex((t) => t.id === id);
    if (i !== -1) todos.splice(i, 1);
  }

  function startLive() {
    live = 0;
    let i = 0;
    const timer = setInterval(() => {
      live = ++i;
      if (i >= 30) clearInterval(timer);
    }, 100);
  }
</script>

<div class="todoapp">
  <h1>todos</h1>
  <div class="entry">
    <input
      id="new-todo"
      type="text"
      placeholder="What needs to be done?"
      bind:value={title}
      onkeydown={onKeydown}
    />
    <button id="add-todo" onclick={addTodo}>Add</button>
  </div>
  <ul id="todo-list">
    {#each visible as t (t.id)}
      <li class={t.done ? 'done' : ''} data-id={t.id}>
        <input type="checkbox" checked={t.done} onchange={() => toggle(t.id)} />
        <span>{t.title}</span>
        <button class="destroy" onclick={() => remove(t.id)}>x</button>
      </li>
    {/each}
  </ul>
  <footer>
    <span id="todo-count">{active} items left</span>
    <button onclick={() => (filter = 'all')}>All</button>
    <button onclick={() => (filter = 'active')}>Active</button>
    <button onclick={() => (filter = 'done')}>Done</button>
  </footer>
  <div class="live">
    <button id="start-live" onclick={startLive}>Start live</button>
    <span id="live-counter">{live}</span>
  </div>
</div>
