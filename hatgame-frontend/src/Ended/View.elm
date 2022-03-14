module Ended.View exposing (html, showResults)

import Containers.Game
import Containers.User
import Html exposing (Html, button, div, h3, label, td, text, th, tr)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
import Msg exposing (Msg)


html : List Containers.User.User -> Maybe Containers.User.User -> List Containers.Game.Team -> List (Html Msg)
html players muser teams =
    [ div [ class "container" ]
        [ div [ class "row" ]
            [ div
                [ classList
                    [ ( "shift-3", True )
                    ]
                ]
                [ div [ class "spacing-both" ]
                    []
                , showResults
                    players
                    muser
                    teams
                , button
                    [ class "btn-primary"
                    , onClick <|
                        Msg.End
                    ]
                    [ text "End" ]
                ]
            ]
        ]
    ]


showResults : List Containers.User.User -> Maybe Containers.User.User -> List Containers.Game.Team -> Html Msg
showResults players muser teams =
    div []
        [ case muser of
            Nothing ->
                h3 [ style "text-align" "center" ] [ text <| "Score" ]

            Just user ->
                h3 [ style "text-align" "center" ] [ text <| Containers.Game.showResult <| Containers.Game.result user teams ]
        , div [ class "spacing-both" ] []
        , div
            [ class "display-window"
            , style "display" "flex"
            , style
                "flex-direction"
                "column"
            , style "justify-content" "space-around"
            ]
          <|
            List.concat [ List.map (showResult players) teams ]
        ]


showResult : List Containers.User.User -> Containers.Game.Team -> Html msg
showResult players team =
    div
        [ style "display" "flex"
        , style "justify-content" "space-around"
        ]
        [ div []
            [ text <| Maybe.withDefault "" (Containers.Game.getUsername players team.playerOne)
            ]
        , div [] [ text <| Maybe.withDefault "" (Containers.Game.getUsername players team.playerTwo) ]
        , div [] [ text <| String.fromInt team.score ]
        ]
