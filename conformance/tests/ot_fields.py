
def test_fields(O):
    errors = []

    O.addVertex("vertex1", "person", {"name": "han", "age": 35, "occupation": "smuggler", "starships": ["millennium falcon"]})

    expected = {
        u"gid": u"vertex1",
        u"label": u"",
        u"data": {u"name": u"han"}
    }

    resp = list(O.query().V().fields(["_gid", "name"]))
    if resp[0] != expected:
        errors.append("vertex contains unexpected fields" % resp)

    return errors
