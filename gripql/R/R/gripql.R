#' @export
gripql <- function(host, user=NULL, password=NULL, token=NULL, credential_file=NULL) {
  env_vars <- Sys.getenv(c("GRIP_USER", "GRIP_PASSWORD", "GRIP_TOKEN", "GRIP_CREDENTIAL_FILE"))
  if (is.null(user)) {
    if (env_vars["GRIP_USER"] != "") {
      user <- env_vars["GRIP_USER"]
    }
  }
  if (is.null(password)) {
    if (env_vars["GRIP_PASSWORD"] != "") {
      password <- env_vars["GRIP_PASSWORD"]
    }
  }
  if (is.null(token)) {
    if (env_vars["GRIP_TOKEN"] != "") {
      token <- env_vars["GRIP_TOKEN"]
    }
  }
  if (is.null(credential_file)) {
    if (env_vars["GRIP_CREDENTIAL_FILE"] != "") {
      credential_file <- env_vars["GRIP_CREDENTIAL_FILE"]
    }
  }
  header <- list("Content-Type" = "application/json",
                 "Accept" = "application/json")
  if (!is.null(token)) {
    header["Authorization"] = sprintf("Bearer %s", token)
  } else if (!(is.null(user) || is.null(password))) {
    header["Authorization"] = sprintf("Basic %s", jsonlite::base64_enc(sprintf("%s:%s", user, password)))
  } else if (!is.null(credential_file)) {
    if (!file.exists(credential_file)) {
      stop("credential file does not exist!")
    }
    creds <- jsonlite::fromJSON(credential_file)
    creds$OauthExpires <- toString(creds$OauthExpires)
    header <- c(header, creds)
  }
  structure(list(), class = "gripql", host = host, header = header)
}

#' @export
print.gripql <- function(x) {
  print(sprintf("host: %s", attr(x, "host")))
}

#' @export
listGraphs <- function(conn) {
  check_class(conn, "gripql")
  response <- httr::GET(url = sprintf("%s/v1/graph", attr(conn, "host")),
                        httr::add_headers(unlist(attr(conn, "header"), use.names = TRUE)),
                        httr::verbose())
  httr::stop_for_status(response)
  if (!grepl("application/json", response$headers$`content-type`)) {
    stop(sprintf("unexpected content-type '%s' in query response",
                 response$headers$`content-type`))
  }
  r <- httr::content(response, as = "parsed", encoding = "UTF-8")
  r$graphs
}

#' @export
graph <- function(conn, graph_name) {
  check_class(conn, "gripql")
  class(conn) <- "gripql.graph"
  attr(conn, "graph") <- graph_name
  conn
}

#' @export
print.gripql.graph <- function(x) {
  print(sprintf("host: %s", attr(x, "host")))
  print(sprintf("graph: %s", attr(x, "graph")))
}

#' @export
getSchema <- function(conn) {
  check_class(conn, "gripql.graph")
  response <- httr::GET(url = sprintf("%s/v1/graph/%s/schema", attr(conn, "host"), attr(conn, "graph")),
                        httr::add_headers(unlist(attr(conn, "header"), use.names = TRUE)),
                        httr::verbose())
  httr::stop_for_status(response)
  if (!grepl("application/json", response$headers$`content-type`)) {
    stop(sprintf("unexpected content-type '%s' in query response",
                 response$headers$`content-type`))
  }
  r <- httr::content(response, as = "parsed", encoding = "UTF-8")
  r
}

#' @export
query <- function(conn) {
  check_class(conn, "gripql.graph")
  class(conn) <- "gripql.graph.query"
  attr(conn, "query") <- list()
  conn
}

#' @export
print.gripql.graph.query <- function(x) {
  print(sprintf("host: %s", attr(x, "host")))
  print(sprintf("graph: %s", attr(x, "graph")))
  print(sprintf("query: %s", to_json(x)))
}

append.gripql.graph.query <- function(x, values, after = length(x)) {
  q <- attr(x, "query")
  after <- length(q)
  q[[after + 1]] <- values
  attr(x, "query") <- q
  x
}

#' @export
to_json <- function(q) {
  check_class(q, "gripql.graph.query")
  query <- attr(q, "query")
  query <- list("query" = query)
  jsonlite::toJSON(query, auto_unbox = T, simplifyVector = F)
}

#' @export
execute <- function(q) {
  check_class(q, "gripql.graph.query")
  body <- to_json(q)
  response <- httr::POST(url = sprintf("%s/v1/graph/%s/query", attr(q, "host"),  attr(q, "graph")),
                         body = body,
                         encode = "json",
                         httr::add_headers(unlist(attr(q, "header"), use.names = TRUE)),
                         httr::verbose())
  httr::stop_for_status(response)
  if (!grepl("application/json", response$headers$`content-type`)) {
    stop(sprintf("unexpected content-type '%s' in query response",
                 response$headers$`content-type`))
  }
  httr::content(response, as="text", encoding = "UTF-8") %>%
    trimws() %>%
    strsplit("\n") %>%
    unlist() %>%
    lapply(function(x) {
        r <- jsonlite::fromJSON(x)
        r <- r$result
        if ("vertex" %in% names(r)) {
          r <- r$vertex
        } else if ("edge" %in% names(r)) {
          r <- r$edge
        } else if ("aggregations" %in% names(r)) {
          r <- r$aggregations$aggregations
        } else if ("selections" %in% names(r)) {
          r <- r$selections$selections
        } else if ("render" %in% names(r)) {
          r <- r$render
        }
        r
    })
}

#' @export
V <- function(q,  ids=NULL) {
  check_class(q, "gripql.graph.query")
  if (length(attr(q, "query")) > 0) {
    stop("V() must be at the beginning of your query")
  }
  ids <- wrap_value(ids)
  names(ids) <- NULL
  append.gripql.graph.query(q, list("v" = ids))
}

#' @export
E <- function(q,  ids=NULL) {
  check_class(q, "gripql.graph.query")
  if (length(attr(q, "query")) > 0) {
    stop("E() must be at the beginning of your query")
  }
  ids <- wrap_value(ids)
  names(ids) <- NULL
  append.gripql.graph.query(q, list("e" = ids))
}

#' @export
in_ <- function(q,  labels=NULL) {
  check_class(q, "gripql.graph.query")
  labels <- wrap_value(labels)
  names(labels) <- NULL
  append.gripql.graph.query(q, list("in" = labels))
}

#' @export
inV <- in_

#' @export
out <- function(q, labels=NULL) {
  check_class(q, "gripql.graph.query")
  labels <- wrap_value(labels)
  names(labels) <- NULL
  append.gripql.graph.query(q, list("out" = labels))
}

#' @export
outV <- out

#' @export
both <- function(q, labels=NULL) {
  check_class(q, "gripql.graph.query")
  labels <- wrap_value(labels)
  names(labels) <- NULL
  append.gripql.graph.query(q, list("both" = labels))
}

#' @export
inE <- function(q, labels=NULL) {
  check_class(q, "gripql.graph.query")
  labels <- wrap_value(labels)
  names(labels) <- NULL
  append.gripql.graph.query(q, list("in_e" = labels))
}

#' @export
outE <- function(q, labels=NULL) {
  check_class(q, "gripql.graph.query")
  labels <- wrap_value(labels)
  names(labels) <- NULL
  append.gripql.graph.query(q, list("out_e" = labels))
}

#' @export
bothE <- function(q, labels=NULL) {
  check_class(q, "gripql.graph.query")
  labels <- wrap_value(labels)
  names(labels) <- NULL
  append.gripql.graph.query(q, list("both_e" = labels))
}

#' @export
has <- function(q, expression) {
  check_class(q, "gripql.graph.query")
  check_class(expression, "list")
  append.gripql.graph.query(q, list("has" = expression))
}

#' @export
hasLabel <- function(q, label) {
  check_class(q, "gripql.graph.query")
  label <- wrap_value(label)
  names(label) <- NULL
  append.gripql.graph.query(q, list("hasLabel" = label))
}

#' @export
hasId <- function(q, id) {
  check_class(q, "gripql.graph.query")
  id <- wrap_value(id)
  names(id) <- NULL
  append.gripql.graph.query(q, list("hasId" = id))
}

#' @export
hasKey <- function(q, key) {
  check_class(q, "gripql.graph.query")
  key <- wrap_value(key)
  names(key) <- NULL
  append.gripql.graph.query(q, list("hasKey" = key))
}

#' @export
fields <- function(q, fields=NULL) {
  check_class(q, "gripql.graph.query")
  fields <- wrap_value(fields)
  names(fields) <- NULL
  append.gripql.graph.query(q, list("fields" = field))
}

#' @export
as_ <- function(q, name) {
  check_class(q, "gripql.graph.query")
  append.gripql.graph.query(q, list("as" = name))
}

#' @export
select <- function(q, marks) {
  check_class(q, "gripql.graph.query")
  marks <- wrap_value(marks)
  names(marks) <- NULL
  append.gripql.graph.query(q, list("select" = list("labels" = marks)))
}

#' @export
limit <- function(q, n) {
  check_class(q, "gripql.graph.query")
  append.gripql.graph.query(q, list("limit" = n))
}

#' @export
skip <- function(q, n) {
  check_class(q, "gripql.graph.query")
  append.gripql.graph.query(q, list("skip" = n))
}

#' @export
range <- function(q, start, stop) {
  check_class(q, "gripql.graph.query")
  append.gripql.graph.query(q, list("range" = list("start" = start, "stop" = stop)))
}

#' @export
count <- function(q) {
  check_class(q, "gripql.graph.query")
  append.gripql.graph.query(q, list("count" = ""))
}

#' @export
distinct <- function(q, props=NULL) {
  check_class(q, "gripql.graph.query")
  props <- wrap_value(props)
  names(props) <- NULL
  append.gripql.graph.query(q, list("distinct" = props))
}

#' @export
render <- function(q, template) {
  check_class(q, "gripql.graph.query")
  append.gripql.graph.query(q, list("render" = template))
}

#' @export
aggregate <- function(q, aggregations) {
  check_class(q, "gripql.graph.query")
  aggregations <- wrap_value(aggregations)
  append.gripql.graph.query(q, list("aggregate" = list("aggregations" = aggregations)))
}

#' @export
match <- function(q, queries) {
  check_class(q, "gripql.graph.query")
  if (length(queries) == 1) {
    queries <- list(queries)
  }
  append.gripql.graph.query(q, list("match", list("queries" = queries)))
}
