package utils

import (
    "github.com/gofiber/fiber/v2"
    
    "github.com/C9b3rD3vi1/forge/config"
)

func SetFlash(c *fiber.Ctx, key, value string) error {
    sess, err := config.Store.Get(c)
    if err != nil {
        return err
    }

    sess.Set("_flash_"+key, value)
    return sess.Save()
}

func GetFlash(c *fiber.Ctx, key string) string {
    sess, err := config.Store.Get(c)
    if err != nil {
        return ""
    }

    flashKey := "_flash_" + key
    val := sess.Get(flashKey)

    if val != nil {
        sess.Delete(flashKey)
        sess.Save()
        return val.(string)
    }

    return ""
}
