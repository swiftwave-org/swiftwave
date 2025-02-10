<template>
  <div>
    <!-- Search and Filter Section -->
    <div class="flex items-center py-4 gap-x-4">
      <Input v-model="searchQuery" placeholder="Search..." class="max-w-sm" />
      <div v-if="filters" class="ml-auto flex items-center gap-x-4">
        <Select
          v-for="(filter, key) in filters"
          :key="key"
          v-model="activeFilters[key]"
          :placeholder="'Filter by ' + key"
        >
          <SelectTrigger class="w-[180px]">
            <SelectValue :placeholder="'Filter by ' + key" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="">All</SelectItem>
              <SelectItem
                v-for="option in filter"
                :key="option"
                :value="option"
              >
                {{ option }}
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </div>
    </div>

    <!-- Table Section -->
    <div class="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead
              v-for="column in columns"
              :key="column.key"
              class="whitespace-nowrap"
              :class="{ 'cursor-pointer': column.sortable }"
              @click="column.sortable && sort(column.key)"
            >
              <div class="flex items-center gap-x-2">
                {{ column.label }}
                <div v-if="column.sortable" class="flex flex-col">
                  <IChevronUp
                    class="h-3 w-3"
                    :class="{
                      'text-foreground':
                        sortConfig.key === column.key &&
                        sortConfig.direction === 'asc',
                      'text-muted-foreground':
                        sortConfig.key !== column.key ||
                        sortConfig.direction !== 'asc',
                    }"
                  />
                  <IChevronDown
                    class="h-3 w-3"
                    :class="{
                      'text-foreground':
                        sortConfig.key === column.key &&
                        sortConfig.direction === 'desc',
                      'text-muted-foreground':
                        sortConfig.key !== column.key ||
                        sortConfig.direction !== 'desc',
                    }"
                  />
                </div>
              </div>
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="(item, index) in paginatedData" :key="index">
            <TableCell
              v-for="column in columns"
              :key="column.key"
              class="whitespace-nowrap"
            >
              <slot
                :name="'cell-' + column.key"
                :value="item[column.key]"
                :row="item"
              >
                {{ item[column.key] }}
              </slot>
            </TableCell>
          </TableRow>
          <TableRow v-if="!filteredData.length">
            <TableCell :colspan="columns.length" class="h-24 text-center">
              No results found.
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>

    <!-- Pagination Section -->
    <div class="flex items-center justify-between py-4">
      <div class="flex items-center gap-x-2 text-sm text-muted-foreground">
        <div>
          Showing {{ startIndex + 1 }} to
          {{ Math.min(endIndex, filteredData.length) }} of
          {{ filteredData.length }} results
        </div>
        <Select v-model="pageSize" class="w-[70px]">
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem
                v-for="size in [5, 10, 20, 50]"
                :key="size"
                :value="size.toString()"
              >
                {{ size }}
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
        <div>per page</div>
      </div>
      <div class="flex items-center gap-x-6 lg:gap-x-8">
        <div
          class="flex w-[100px] items-center justify-center text-sm font-medium"
        >
          Page {{ currentPage }} of {{ totalPages }}
        </div>
        <div class="flex items-center gap-x-2">
          <Button
            variant="outline"
            :disabled="currentPage === 1"
            @click="currentPage--"
          >
            Previous
          </Button>
          <Button
            variant="outline"
            :disabled="currentPage === totalPages"
            @click="currentPage++"
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Column {
  key: string;
  label: string;
  sortable?: boolean;
}

interface Props {
  columns: Column[];
  data: any[];
  filters?: Record<string, string[]>;
}

const props = defineProps<Props>();
const searchQuery = ref("");
const currentPage = ref(1);
const pageSize = ref("10");
const activeFilters = ref<Record<string, string>>({});
const sortConfig = ref({
  key: "",
  direction: "" as "asc" | "desc" | "",
});

// Computed properties for filtering and sorting
const filteredData = computed(() => {
  let result = [...props.data];

  // Apply search filter
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    result = result.filter((item) =>
      Object.values(item).some((val) =>
        String(val).toLowerCase().includes(query)
      )
    );
  }

  // Apply column filters
  Object.entries(activeFilters.value).forEach(([key, value]) => {
    if (value) {
      result = result.filter((item) => String(item[key]) === value);
    }
  });

  // Apply sorting
  if (sortConfig.value.key && sortConfig.value.direction) {
    result.sort((a, b) => {
      const aVal = a[sortConfig.value.key];
      const bVal = b[sortConfig.value.key];
      const modifier = sortConfig.value.direction === "asc" ? 1 : -1;

      if (typeof aVal === "string" && typeof bVal === "string") {
        return modifier * aVal.localeCompare(bVal);
      }
      return modifier * (aVal - bVal);
    });
  }

  return result;
});

// Pagination
const totalPages = computed(() =>
  Math.ceil(filteredData.value.length / parseInt(pageSize.value))
);

const startIndex = computed(
  () => (currentPage.value - 1) * parseInt(pageSize.value)
);
const endIndex = computed(() => startIndex.value + parseInt(pageSize.value));

const paginatedData = computed(() =>
  filteredData.value.slice(startIndex.value, endIndex.value)
);

// Watch for changes that should reset pagination
watch([searchQuery, activeFilters, pageSize], () => {
  currentPage.value = 1;
});

// Sorting function
function sort(key: string) {
  if (sortConfig.value.key === key) {
    if (sortConfig.value.direction === "asc") {
      sortConfig.value.direction = "desc";
    } else if (sortConfig.value.direction === "desc") {
      sortConfig.value.direction = "";
      sortConfig.value.key = "";
    } else {
      sortConfig.value.direction = "asc";
    }
  } else {
    sortConfig.value.key = key;
    sortConfig.value.direction = "asc";
  }
}
</script>
