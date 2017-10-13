local data = {}
local n = 1000 --should be an even number as we use n/2 below...

for i=1,n do
    data[i] = tostring(i)
end

function testSimpleAdd()
    local bs, bg
    bg = parsejoy.makeStringBitGrammar()
    bs = parsejoy.StringBitSet(bg)
    luaunit.assertEquals(bs:Contains("test"),false)
    bs:Add("test")
    luaunit.assertEquals(bs:Contains("test"),true)
    luaunit.assertEquals(bg:Get("test"),1)
end

function testCount()
    local bs, bg
    bg = parsejoy.makeStringBitGrammar()
    bs = parsejoy.StringBitSet(bg)
    bs:Add("test")
    luaunit.assertEquals(bs:N(),1)
end

function testRemove()
    local bs, bg
    bg = parsejoy.makeStringBitGrammar()
    bs = parsejoy.StringBitSet(bg)
    bs:Add("test")
    bs:Add("tests")
    luaunit.assertEquals(bs:Contains("tests"),true)
    luaunit.assertEquals(bs:Contains("test"),true)
    bs:Remove("test")
    luaunit.assertEquals(bs:Contains("tests"),true)
    luaunit.assertEquals(bs:Contains("test"),false)
    bs:Remove("tests")
    luaunit.assertEquals(bs:Contains("tests"),false)
    luaunit.assertEquals(bs:Contains("test"),false)
end

function testIds()
    local bs, bg, testId
    bg = parsejoy.makeStringBitGrammar()
    bs = parsejoy.StringBitSet(bg)
    testId = bg:GetOrAdd("test")
    bs:Add("test")
    luaunit.assertEquals(bs:ContainsId(testId),true)
end

function testSubtract()
    local bg, bs1, bs2, bs3
    bg = parsejoy.makeStringBitGrammar()
    bs1 = parsejoy.StringBitSet(bg)
    bs2 = parsejoy.StringBitSet(bg)
    bs1:Add("a")
    bs1:Add("b")
    bs1:Add("c")
    bs2:Add("c")
    bs3 = bs1:Subtract(bs2)
    luaunit.assertEquals(bs3:N(),2)
    luaunit.assertEquals(bs3:Contains("c"),false)
    luaunit.assertEquals(bs3:Contains("a"),true)
    luaunit.assertEquals(bs3:Contains("b"),true)
end

function testLargeSubtract()
    local bg, bs1, bs2, bs3
    bg = parsejoy.makeStringBitGrammar()
    bs1 = parsejoy.StringBitSet(bg)
    bs2 = parsejoy.StringBitSet(bg)
    for key, value in pairs(data) do
        if key % 2 == 0 then
            bs1:Add(value)
            luaunit.assertEquals(bs1:N(),key/2)
        end
        bs2:Add(value)
        luaunit.assertEquals(bs2:N(),key)
    end
    luaunit.assertEquals(bs1:N(),n/2)
    luaunit.assertEquals(bs2:N(),n)

    bs3 = bs2:Subtract(bs1)
    luaunit.assertEquals(bs3:N(),n/2)

    --we make sure bs3 contains only values that are in bs2 but not in bs1
    for key, value in pairs(data) do
        if key % 2 == 0 then
            luaunit.assertEquals(bs3:Contains(value),false)
        else
            luaunit.assertEquals(bs3:Contains(value),true)
        end
    end
end

function testUnion()
    local bg, bs1, bs2, bs3
    bg = parsejoy.makeStringBitGrammar()
    bs1 = parsejoy.StringBitSet(bg)
    bs2 = parsejoy.StringBitSet(bg)
    bs1:Add("a")
    bs1:Add("b")
    bs1:Add("c")
    bs2:Add("c")
    bs2:Add("d")
    bs2:Add("e")
    bs3 = bs1:Union(bs2)
    luaunit.assertEquals(bs3:N(),5)
    luaunit.assertEquals(bs3:Contains("a"),true)
    luaunit.assertEquals(bs3:Contains("b"),true)
    luaunit.assertEquals(bs3:Contains("c"),true)
    luaunit.assertEquals(bs3:Contains("d"),true)
    luaunit.assertEquals(bs3:Contains("e"),true)
end

function testLargeUnion()
    local bg, bs1, bs2, bs3
    bg = parsejoy.makeStringBitGrammar()
    bs1 = parsejoy.StringBitSet(bg)
    bs2 = parsejoy.StringBitSet(bg)
    for key, value in pairs(data) do
        if key % 2 == 0 then
            bs1:Add(value)
            luaunit.assertEquals(bs1:N(),key/2)
        end
        bs2:Add(value)
        luaunit.assertEquals(bs2:N(),key)
    end
    luaunit.assertEquals(bs1:N(),n/2)
    luaunit.assertEquals(bs2:N(),n)

    bs3 = bs2:Union(bs1)
    luaunit.assertEquals(bs3:N(),n)

    for key, value in pairs(data) do
        luaunit.assertEquals(bs3:Contains(value),true)
    end
end

function testIntersect()
    local bg, bs1, bs2, bs3
    bg = parsejoy.makeStringBitGrammar()
    bs1 = parsejoy.StringBitSet(bg)
    bs2 = parsejoy.StringBitSet(bg)
    bs1:Add("a")
    bs1:Add("b")
    bs1:Add("c")
    bs2:Add("c")
    bs2:Add("d")
    bs2:Add("e")
    bs3 = bs1:Intersect(bs2)
    luaunit.assertEquals(bs3:N(),1)
    luaunit.assertEquals(bs3:Contains("a"),false)
    luaunit.assertEquals(bs3:Contains("b"),false)
    luaunit.assertEquals(bs3:Contains("c"),true)
    luaunit.assertEquals(bs3:Contains("d"),false)
    luaunit.assertEquals(bs3:Contains("e"),false)
end

function testLargeIntersect()
    local bg, bs1, bs2, bs3
    bg = parsejoy.makeStringBitGrammar()
    bs1 = parsejoy.StringBitSet(bg)
    bs2 = parsejoy.StringBitSet(bg)
    for key, value in pairs(data) do
        if key % 2 == 0 then
            bs1:Add(value)
            luaunit.assertEquals(bs1:N(),key/2)
        end
        bs2:Add(value)
        luaunit.assertEquals(bs2:N(),key)
    end
    luaunit.assertEquals(bs1:N(),n/2)
    luaunit.assertEquals(bs2:N(),n)

    bs3 = bs2:Intersect(bs1)
    luaunit.assertEquals(bs3:N(),n/2)

    for key, value in pairs(data) do
        if bs2:Contains(value) and bs1:Contains(value) then
            luaunit.assertEquals(bs3:Contains(value),true)
        else
            luaunit.assertEquals(bs3:Contains(value),false)
        end
    end
end