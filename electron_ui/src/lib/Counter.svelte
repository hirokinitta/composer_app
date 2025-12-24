<!-- <script>
  let count = $state(0)
  const increment = () => {
    count += 1
  }
</script>

<button onclick={increment}>
  count is {count}
</button> -->

<script>
// @ts-nocheck

    // 正しい変数名は 'count'
    let count = 0;
</script>

<button on:click={count++}>
    Clicks: {count}
</button>

<style>
    /* ... スタイルはそのまま ... */
</style>
