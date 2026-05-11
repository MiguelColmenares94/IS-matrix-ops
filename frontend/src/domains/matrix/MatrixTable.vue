<template>
  <div class="overflow-x-auto">
    <table class="text-sm border-collapse">
      <tbody>
        <tr v-for="(row, i) in data" :key="i">
          <td
            v-for="(val, j) in row"
            :key="j"
            class="border border-brand-lavender px-3 py-1 text-right font-mono text-brand-navy"
          >{{ val.toFixed(6) }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
defineProps({ data: Array })
</script>
