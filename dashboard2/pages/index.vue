<template>
  <div class="container py-10">
    {{ authStore.sessionRelativeTimeoutStatus }}
    <select v-model="$colorMode.preference">
      <option value="system">System</option>
      <option value="light">Light</option>
      <option value="dark">Dark</option>
      <option value="sepia">Sepia</option>
    </select>
    <div v-if="applicationsResult">
      <DataTable
        :columns="columns"
        :data="applicationsResult.applications"
        :filters="tableFilters"
      >
        <template #cell-name="{ row }">
          <div class="flex items-center gap-2">
            <Badge v-if="row.isSleeping" variant="secondary">Sleeping</Badge>
            {{ row.name }}
          </div>
        </template>

        <template #cell-status="{ row }">
          <Badge
            :variant="
              row.realtimeInfo.HealthStatus === 'healthy'
                ? 'default'
                : 'destructive'
            "
          >
            {{ row.realtimeInfo.HealthStatus }}
          </Badge>
        </template>

        <template #cell-replicas="{ row }">
          {{ row.realtimeInfo.RunningReplicas }}/{{
            row.realtimeInfo.DesiredReplicas
          }}
        </template>

        <template #cell-lastDeployed="{ row }">
          <time
            :datetime="row.latestDeployment.createdAt"
            class="text-sm text-muted-foreground"
          >
            {{ new Date(row.latestDeployment.createdAt).toLocaleDateString() }}
          </time>
        </template>
      </DataTable>
    </div>

    <div v-if="isApplicationsLoading" class="flex justify-center py-4">
      <ILoaderCircle class="h-6 w-6 animate-spin" />
    </div>
  </div>
</template>

<script setup lang="ts">
const authStore = useAuthStore();
const colormode = useColorMode();

const columns = [
  { key: "name", label: "Name", sortable: true },
  { key: "status", label: "Status", sortable: true },
  { key: "replicas", label: "Replicas" },
  { key: "deploymentMode", label: "Deployment Mode", sortable: true },
  { key: "lastDeployed", label: "Last Deployed", sortable: true },
  { key: "source", label: "Source", sortable: true },
];

const tableFilters = {
  deploymentMode: ["replicated", "global"],
  source: ["git", "registry"],
};

const {
  result: applicationsResult,
  refetch: refetchApplications,
  loading: isApplicationsLoading,
  onError: onApplicationsError,
} = useQuery(
  gql`
    query {
      applications(includeGroupedApplications: false) {
        id
        name
        replicas
        isSleeping
        realtimeInfo {
          InfoFound
          DesiredReplicas
          RunningReplicas
          DeploymentMode
          HealthStatus
        }
        latestDeployment {
          status
          upstreamType
          createdAt
        }
      }
    }
  `,
  null,
  {
    pollInterval: 60000,
  }
);

onApplicationsError((err) => {
  // toast.error(err.message)
  console.error(err);
});
</script>

<style scoped></style>
