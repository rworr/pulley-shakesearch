let page;

const Controller = {

  newSearch: (ev) => {
    ev.preventDefault();
    page = 1;
    Controller.search();
  },

  loadMore: (ev) => {
    ev.preventDefault();
    page += 1;
    Controller.search();
  },

  search: () => {
    const form = document.getElementById("form");
    const data = Object.fromEntries(new FormData(form));
    const response = fetch(`/search?q=${data.query}&p=${page}`).then((response) => {
      response.json().then((results) => {
        Controller.updateTable(results);
      });
    });
  },

  updateTable: (results) => {
    const table = document.getElementById("table-body");
    const rows = (page > 1) ? [table.innerHTML] : [];
    for (let result of results) {
      rows.push(`<tr><td>${result}</td></tr>`);
    }
    table.innerHTML = rows;
  },
};

const form = document.getElementById("form");
form.addEventListener("submit", Controller.newSearch);

const loadMore = document.getElementById("load-more")
loadMore.addEventListener("click", Controller.loadMore);
