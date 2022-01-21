module Home.View exposing (html)

import Containers.Statistics
import Containers.User
import Home
import Html exposing (Html, br, button, div, input, label, text)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput)
import Html.Utils
import Maybe.Utils
import Msg exposing (Msg)


html : Maybe Containers.User.WithToken -> Maybe String -> Home.Data -> Html Msg
html tokenUser err homeData =
    div []
        [ div [ class "header-home" ]
            [ label [ class "user-label" ] [ text <| Containers.User.maybeUsernameWithToken tokenUser ]
            , button
                [ class "btn-secondary"
                , onClick (Msg.ToChangePage homeData)
                ]
                [ text "Edit Profile" ]
            ]
        , div [ class "container" ]
            [ div [ class "row" ]
                [ div
                    [ classList
                        [ ( "shift-3", True )
                        , ( "has-error", err /= Nothing )
                        ]
                    ]
                    [ div [ class "spacing-both" ] []
                    , button
                        [ class "btn-host"
                        , onClick Msg.Create
                        ]
                        [ text "Host" ]
                    , div [ class "spacing-both" ] []
                    , div
                        [ class "form-group" ]
                        [ label [ class "control-label" ] [ text "Game" ]
                        , input
                            [ class "form-control"
                            , type_ "text"
                            , value <|
                                case homeData.gameId of
                                    Nothing ->
                                        ""

                                    Just n ->
                                        String.fromInt n
                            , onInput <|
                                \str ->
                                    Msg.ChangeHome <|
                                        case Maybe.Utils.toMaybeInt str of
                                            Maybe.Utils.NoString ->
                                                { homeData
                                                    | gameId = Nothing
                                                }

                                            Maybe.Utils.NotInt ->
                                                homeData

                                            Maybe.Utils.ParsedInt n ->
                                                { homeData
                                                    | gameId = Just n
                                                }
                            ]
                            []
                        , Html.Utils.divOnJust err
                        ]
                    , button
                        [ class "btn-secondary"
                        , onClick (Msg.CheckGame homeData)
                        ]
                        [ text "Join" ]
                    , div [ class "spacing-both" ] []
                    , Html.Utils.when (homeData.stats /= Nothing) <| Containers.Statistics.show homeData.stats
                    ]
                ]
            ]
        ]
