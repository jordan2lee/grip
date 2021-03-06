from __future__ import absolute_import


def test_render(man):
    errors = []

    G = man.setGraph("swapi")

    query = G.query().V().hasLabel("Person").render(
        {
            "Name": "name",
            "Age": "age"
        }
    )
    for row in query:
        if 'Age' not in row or "Name" not in row:
            errors.append("Missing fields")

    query = G.query().V().hasLabel("Person").render(
        {
            "Name": "name",
            "NonExistent": "non-existent"
        }
    )
    for row in query:
        if 'NonExistent' not in row or "Name" not in row:
            errors.append("Missing fields")

    query = G.query().V().hasLabel("Person").render(["name", "age"])
    for row in query:
        if not isinstance(row, list):
            errors.append("unexpected output format")
        if len(row) != 2:
            errors.append("Missing fields")

    query = G.query().V().hasLabel("Person").render(["name", "non-existent"])
    for row in query:
        if not isinstance(row, list):
            errors.append("unexpected output format")
        if len(row) != 2:
            errors.append("Missing fields")

    return errors
