module Api.{{.ModuleName}} exposing (create, delete, getById, list, update)

import Api.Endpoint exposing (Endpoint, url)

create =
    url [ {{if .Prefix }} "{{.Prefix}}", {{end}} "{{.Endpoint}}" ] []

delete id =
    url [ {{if .Prefix }} "{{.Prefix}}", {{end}} "{{.Endpoint}}", String.fromInt id ] []

getById id =
    url [ {{if .Prefix }} "{{.Prefix}}", {{end}} "{{.Endpoint}}", String.fromInt id ] []

list =
    url [ {{if .Prefix }} "{{.Prefix}}", {{end}} "{{.Endpoint}}" ] []

update id =
    url [ {{if .Prefix }} "{{.Prefix}}", {{end}} "{{.Endpoint}}", String.fromInt id ] []