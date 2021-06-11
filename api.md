
###### Data structures with references

GroceryItem { id:... name:... current: ... minimum: ... }

ListItem { groceryItemId: ... quantity: ... checked: ... position: ... }

List { id: ... title: ... items: [ ... ] (ListItem) }


###### REST API

Base Path: /api/

Lists:

    GET /lists
        includeItems default false [ { id: ..., title: ... } ]
        GET /lists?includeItems=true [ { id: ..., title: ..., items: [ ... ] } ]
    GET /lists/:id { id: ..., title, ..., items: [ ... ] }
    POST /lists { title: ... } => { id: ..., title: ... }
    PATCH /lists/:id { title: ... } | { items: [ ... ] } | { title: ..., items: [ ... ] }
    PATCH /lists/:id/:groceryItemId { checked ... } | ...
    DELETE /lists/:id

GroceryItems:

    GET /items
    POST /items
    PATCH /items/:id
    DELETE /items/:id (hierbei muss der Server auch alle ListItems, die das GroceryItem referenzieren, löschen)
