local delimiters = {"|", "!", "+", ";", "~", "@", "#", "%", "^", ":"}

local function wrap_code(code)
  if FORMAT ~= "latex" and FORMAT ~= "beamer" then
    return nil
  end

  local text = code.text or ""
  for _, delim in ipairs(delimiters) do
    if not string.find(text, delim, 1, true) then
      return pandoc.RawInline("latex", "\\path" .. delim .. text .. delim)
    end
  end

  return nil
end

function Table(tbl)
  return tbl:walk({
    Code = wrap_code,
  })
end
