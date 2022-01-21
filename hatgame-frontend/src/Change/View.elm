module Change.View exposing (html)

import Html exposing (Html, button, div, input, label, text)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput)
import Html.Utils
import Maybe.Utils
import Msg exposing (Msg)
import UserCredentials



-- html : UserCredentials.Data -> Html Msg
-- html changeData =
--     div []
--         [ div []
--             [ text (Maybe.withDefault "" changeData.email)
--             ]
--         , br [] []
--         , input
--             [ type_ "password"
--             , placeholder "new password"
--             , value (Maybe.withDefault "" changeData.password)
--             , onInput <|
--                 \str ->
--                     Msg.ChangeChangePage
--                         { changeData
--                             | password = Maybe.Utils.fromString str
--                         }
--             ]
--             []
--         , br [] []
--         , input
--             [ placeholder "username"
--             , value (Maybe.withDefault "" changeData.username)
--             , onInput <|
--                 \str ->
--                     Msg.ChangeChangePage
--                         { changeData
--                             | username = Maybe.Utils.fromString str
--                         }
--             ]
--             []
--         , br [] []
--         , button [ onClick Msg.ChangeUserData ] [ text "Apply" ]
--         ]


html : Maybe String -> UserCredentials.Data -> Html Msg
html err changeData =
    div [ class "container" ]
        [ div [ class "row" ]
            [ div
                [ classList
                    [ ( "shift-3", True )
                    , ( "has-error", err /= Nothing )
                    ]
                ]
                [ div [ class "spacing-both" ] []
                , div
                    [ classList
                        [ ( "form-group", True )
                        , ( "has-error", changeData.email == Nothing )
                        ]
                    ]
                    [ label [ class "control-label" ] [ text "Email" ]
                    , input
                        [ class "form-control"
                        , type_ "text"
                        , readonly True
                        , value (Maybe.withDefault "" changeData.email)
                        , onInput <|
                            \str ->
                                Msg.ChangeRegister
                                    { changeData
                                        | email = Maybe.Utils.fromString str
                                    }
                        ]
                        []
                    ]
                , div
                    [ classList
                        [ ( "form-group", True )
                        , ( "has-error", changeData.password == Nothing )
                        ]
                    ]
                    [ label [ class "control-label" ] [ text "Password" ]
                    , input
                        [ class "form-control"
                        , type_ "password"
                        , value (Maybe.withDefault "" changeData.password)
                        , onInput <|
                            \str ->
                                Msg.ChangeChangePage
                                    { changeData
                                        | password = Maybe.Utils.fromString str
                                    }
                        ]
                        []
                    ]
                , div
                    [ class <|
                        "form-group"
                    ]
                    [ label [ class "control-label" ] [ text "Username" ]
                    , input
                        [ class "form-control"
                        , type_ "text"
                        , value (Maybe.withDefault "" changeData.username)
                        , onInput <|
                            \str ->
                                Msg.ChangeChangePage
                                    { changeData
                                        | username = Maybe.Utils.fromString str
                                    }
                        ]
                        []
                    ]
                , button
                    [ class "btn-primary"
                    , onClick Msg.ChangeUserData
                    , disabled <| changeData.password == Nothing
                    ]
                    [ text "Apply" ]
                , Html.Utils.divOnJust err
                ]
            ]
        ]
