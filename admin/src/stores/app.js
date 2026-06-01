import { defineStore } from 'pinia';
import { ref } from 'vue';
export const useAppStore = defineStore('app', () => {
    const sidebarCollapsed = ref(false);
    const tags = ref([]);
    const breadcrumbs = ref([]);
    function toggleSidebar() {
        sidebarCollapsed.value = !sidebarCollapsed.value;
    }
    function addTag(tag) {
        const exists = tags.value.find((t) => t.path === tag.path);
        if (!exists) {
            tags.value.push(tag);
        }
    }
    function removeTag(path) {
        const idx = tags.value.findIndex((t) => t.path === path);
        if (idx > -1) {
            tags.value.splice(idx, 1);
        }
    }
    function removeOtherTags(path) {
        tags.value = tags.value.filter((t) => t.path === path);
    }
    function removeAllTags() {
        tags.value = [];
    }
    function setBreadcrumbs(items) {
        breadcrumbs.value = items;
    }
    return {
        sidebarCollapsed,
        tags,
        breadcrumbs,
        toggleSidebar,
        addTag,
        removeTag,
        removeOtherTags,
        removeAllTags,
        setBreadcrumbs,
    };
});
