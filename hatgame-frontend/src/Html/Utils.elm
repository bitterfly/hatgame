module Html.Utils exposing (divOnJust, printOnJust, printOnNothing, when, whenMaybe)

import Html exposing (Html, div, text)
import Html.Attributes exposing (..)


when : Bool -> Html msg -> Html msg
when t html =
    if t then
        html

    else
        text ""


whenMaybe : Maybe String -> Html msg
whenMaybe =
    text << Maybe.withDefault "default"


printOnJust : Maybe String -> String -> String
printOnJust mstr str =
    if mstr /= Nothing then
        str

    else
        ""


printOnNothing : Maybe String -> String -> String
printOnNothing mstr str =
    if mstr == Nothing then
        str

    else
        ""


opacityMaybe : Maybe String -> String
opacityMaybe str =
    if str == Nothing then
        "0"

    else
        "100"


divOnJust : Maybe String -> Html msg
divOnJust str =
    div
        [ style "opacity" <|
            opacityMaybe str
        , class
            "help-block"
        ]
        [ whenMaybe str ]
