import { defineStore } from 'pinia';

export const locaShareCount = defineStore('localShareCount', {
  state: () => {
    return {
      count: "0"
    };
  },
  getters: {
    getLoalShareCoount: (state) => {
      return state.count
    },
  },
  actions: {
    set(n) {
      this.count = String(n);
    },
  }
});