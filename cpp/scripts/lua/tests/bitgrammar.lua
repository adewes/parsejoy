function testAddAs()
    local bg = parsejoy.StringBitGrammar()
    bg:AddAs("test",1)
    luaunit.assertEquals(bg:Get("test"),1)
    luaunit.assertEquals(bg:AddAs("test",1),nil)
    local res, ok = pcall(function()bg:AddAs("tests",1)end)
    luaunit.assertEquals(res, false)
end
