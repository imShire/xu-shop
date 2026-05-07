import { ref, watch } from 'vue';
export function useTable(fetchFn) {
    const list = ref([]);
    const total = ref(0);
    const page = ref(1);
    const pageSize = ref(20);
    const loading = ref(false);
    const fetch = async (extraParams = {}) => {
        loading.value = true;
        try {
            const res = await fetchFn({
                page: page.value,
                page_size: pageSize.value,
                ...extraParams,
            });
            list.value = res.list;
            total.value = res.total;
        }
        finally {
            loading.value = false;
        }
    };
    watch([page, pageSize], () => {
        fetch();
    });
    return { list, total, page, pageSize, loading, fetch };
}
