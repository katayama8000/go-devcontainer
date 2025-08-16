```bash
curl -g 'http://localhost:8080/graphql?query={todos{id,title,completed}}'

curl -g 'http://localhost:8080/graphql?query={todo(id:2){title}}'

curl -X POST \
-H "Content-Type: application/json" \
-d '{"query": "{ todos { id, title } }"}' \
http://localhost:8080/graphql

curl -X POST \
-H "Content-Type: application/json" \
-d '{"query": "mutation { updateTodo(id: 1, completed: true) { id, title, completed } }"}' \
http://localhost:8080/graphql
```