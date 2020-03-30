module Data.{{ .ModuleName }} exposing
    ({{ .ModuleName }}
    , create
    , destroy
    , getById
    , list
    , update
    )

import Api.{{ .APIModuleName }} as Endpoints
import Http
import Iso8601
import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline exposing (required, optional)
import Json.Encode as Encode
import Json.Encode.Extra as Extra
import Session exposing (Session, delete, get, post, put)
import Time

-- TYPES

type alias {{ .ModuleName }} =
    { {{ .PKName }} : {{ .PKType }}
    {{ range .Fields }}, {{.Name}} : {{if .Nullable}}Maybe {{end}}{{.Type}}
    {{end}}
    }


-- ENCODERS/DECODERS

decoder : Decoder {{ .ModuleName }}
decoder =
    Decode.succeed {{ .ModuleName }}
        {{range .DecodeFields}}|> {{if .Nullable}}optional{{else}}required{{end}} "{{.Name}}" Decode.{{.Type}} {{if .Nullable}}Nothing{{end}}
        {{end}}


encode : {{.ModuleName}} -> Encode.Value
encode x =
    Encode.object
        [ ("{{.PKName}}", Encode.{{.PKEncodeType}} x.{{.PKName}})
        {{range .EncodeFields}}, ("{{.Name}}", {{if .Nullable}}Extra.maybe {{end}}Encode.{{.Type}} x.{{.Name}})
        {{end}}
        ]


-- API ACCESS

create : Session -> {{.ModuleName}} -> (Result Http.Error {{.PKType}} -> msg) -> Cmd msg
create session x msg =
    post
        session
        Endpoints.create
        (Http.jsonBody (encode x))
        (Http.expectJson msg Decode.{{.PKEncodeType}})

list : Session -> (Result Http.Error (List {{.ModuleName}}) -> msg) -> Cmd msg
list session msg =
    get
        session
        Endpoints.list
        (Http.expectJson msg (Decode.list decoder))

getById : Session -> {{.PKType}} -> (Result Http.Error {{.ModuleName}} -> msg) -> Cmd msg
getById session id msg =
    get
        session
        (Endpoints.getById id)
        (Http.expectJson msg decoder)

update : Session -> ({{.PKType}}, {{.ModuleName}}) -> (Result Http.Error Bool -> msg) -> Cmd msg
update session (id, x) msg =
    put
        session
        (Endpoints.update id)
        (Http.jsonBody (encode x))
        (Http.expectJson msg Decode.bool)

destroy : Session -> {{.PKType}} -> (Result Http.Error Bool -> msg) -> Cmd msg
destroy session id msg =
    delete
        session
        (Endpoints.delete id)
        (Http.expectJson msg Decode.bool)
