module Api.Endpoint exposing (Endpoint, unwrap, url)

import Url.Builder exposing (QueryParameter, int)



-- TYPES


type Endpoint
    = Endpoint String


unwrap : Endpoint -> String
unwrap (Endpoint str) =
    str


url : List String -> List QueryParameter -> Endpoint
url paths params =
    Url.Builder.crossOrigin "http://localhost:8080" paths params
        |> Endpoint







