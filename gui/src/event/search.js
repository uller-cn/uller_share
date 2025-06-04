import { reactive } from "vue";

// const searchTitle = ref('');
// const selectExt = ref([]);

// export const useSearchTitle = () => {
//   return { searchTitle };
// };

// export const useSelectExt = () => {
//     return { selectExt };
//   };

// export const setSearch = (newSearchTitle,newSelectExt) => {
//   searchTitle.value = newSearchTitle;
//   selectExt.value = newSelectExt;
// };

const searchBus = reactive({
  searchClick: (searchTitle, selectExt) => {
    console.log("searchBus",searchTitle, selectExt)
    return { searchTitle, selectExt };
  },
  searchReceiveClick: () => {
    console.log("searchBus",searchTitle, selectExt)
    return { searchTitle, selectExt };
  },
});

export default searchBus;
