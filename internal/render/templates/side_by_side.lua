-- md2pdf side-by-side layout filter
-- Supports two paired panes rendered as LaTeX minipages or HTML flex items.

local function trim(text)
  if not text then
    return nil
  end
  return (text:gsub("^%s+", ""):gsub("%s+$", ""))
end

local function has_class(el, class_name)
  return el.classes and el.classes:includes(class_name)
end

local function parse_ratio(attr)
  if not attr or not attr.attributes then
    return nil, nil
  end
  local ratio = trim(attr.attributes["ratio"])
  if not ratio then
    return nil, nil
  end
  local left, right = ratio:match("^(%d*%.?%d+)%s*:%s*(%d*%.?%d+)$")
  if not left or not right then
    return nil, nil
  end
  left = tonumber(left)
  right = tonumber(right)
  if not left or not right or left <= 0 or right <= 0 then
    return nil, nil
  end
  local sum = left + right
  return left / sum, right / sum
end

local function parse_fraction(text)
  text = trim(text)
  if not text then
    return nil
  end
  local pct = text:match("^(%d*%.?%d+)%%$")
  if pct then
    local value = tonumber(pct)
    if value and value > 0 then
      return value / 100
    end
    return nil
  end
  local value = tonumber(text)
  if not value or value <= 0 then
    return nil
  end
  if value <= 1 then
    return value
  end
  if value <= 100 then
    return value / 100
  end
  return nil
end

local function resolve_widths(attr)
  local left_ratio, right_ratio = parse_ratio(attr)
  if left_ratio and right_ratio then
    return left_ratio, right_ratio
  end

  local left = parse_fraction(attr.attributes["left"] or attr.attributes["left-width"])
  local right = parse_fraction(attr.attributes["right"] or attr.attributes["right-width"])

  if left and right then
    local sum = left + right
    if sum > 1 then
      return left / sum, right / sum
    end
    return left, right
  end
  if left then
    if left >= 1 then
      return 0.5, 0.5
    end
    return left, 1 - left
  end
  if right then
    if right >= 1 then
      return 0.5, 0.5
    end
    return 1 - right, right
  end
  return 0.5, 0.5
end

local function parse_length(text)
  text = trim(text)
  if not text then
    return nil, nil
  end
  local value, unit = text:match("^(%-?%d*%.?%d+)%s*(pt|mm|cm|in|ex|em)$")
  if not value or not unit then
    return nil, nil
  end
  return tonumber(value), unit
end

local function format_decimal(value)
  local text = string.format("%.4f", value)
  text = text:gsub("0+$", ""):gsub("%.$", "")
  if text == "-0" or text == "" then
    return "0"
  end
  return text
end

local function resolve_gap(attr)
  local gap = nil
  if attr and attr.attributes then
    gap = trim(attr.attributes["gap"] or attr.attributes["column-gap"])
  end
  if not gap or gap == "" then
    gap = "2em"
  end
  return gap
end

local function width_expr(fraction, gap)
  local gap_value, gap_unit = parse_length(gap)
  local fraction_text = format_decimal(fraction)
  if gap_value and gap_unit then
    local shared_gap = format_decimal(fraction * gap_value)
    return "\\dimexpr " .. fraction_text .. "\\linewidth - " .. shared_gap .. gap_unit .. "\\relax"
  end
  return fraction_text .. "\\linewidth"
end

local function css_width_expr(fraction, gap)
  local gap_value, gap_unit = parse_length(gap)
  local pct = format_decimal(fraction * 100)
  if gap_value and gap_unit then
    local shared_gap = format_decimal(fraction * gap_value)
    return "calc(" .. pct .. "% - " .. shared_gap .. gap_unit .. ")"
  end
  return pct .. "%"
end

local function resolve_valign(attr)
  local valign = "top"
  if attr and attr.attributes and attr.attributes["valign"] then
    valign = trim(attr.attributes["valign"]):lower()
  end
  if valign == "center" or valign == "middle" or valign == "c" then
    return "c", "center"
  end
  if valign == "bottom" or valign == "b" then
    return "b", "flex-end"
  end
  return "t", "flex-start"
end

local function resolve_align(attr)
  local align = ""
  if attr and attr.attributes and attr.attributes["align"] then
    align = trim(attr.attributes["align"]):lower()
  end
  if align == "center" then
    return "\\centering", "center"
  end
  if align == "right" then
    return "\\raggedleft", "right"
  end
  return "\\raggedright", "left"
end

local function flatten_figures(blocks)
  local wrapper = pandoc.Div(blocks)
  wrapper = pandoc.walk_block(wrapper, {
    Figure = function(el)
      return el.content
    end,
  })
  return wrapper.content
end

local function collect_panes(element)
  local divs = pandoc.List:new()
  local left = nil
  local right = nil

  for _, block in ipairs(element.content) do
    if block.t == "Div" then
      divs:insert(block)
      if not left and (has_class(block, "left") or has_class(block, "first")) then
        left = block
      elseif not right and (has_class(block, "right") or has_class(block, "second")) then
        right = block
      end
    end
  end

  if not left and divs[1] then
    left = divs[1]
  end
  if not right and divs[2] then
    right = divs[2]
  end

  if not left or not right then
    return nil, nil
  end

  return left, right
end

local function format_latex(element)
  local left, right = collect_panes(element)
  if not left or not right then
    return nil
  end

  local left_fraction, right_fraction = resolve_widths(element.attr)
  local gap = resolve_gap(element.attr)
  local valign_latex = resolve_valign(element.attr)
  local align_latex = resolve_align(element.attr)

  local blocks = pandoc.List:new()
  blocks:insert(pandoc.RawBlock("latex",
    "\\par\\noindent\\begin{minipage}[" .. valign_latex .. "]{" .. width_expr(left_fraction, gap) .. "}\n" ..
    align_latex .. "\n"))
  blocks:extend(flatten_figures(left.content))
  blocks:insert(pandoc.RawBlock("latex",
    "\\end{minipage}\\hspace{" .. gap .. "}\\begin{minipage}[" .. valign_latex .. "]{" .. width_expr(right_fraction, gap) .. "}\n" ..
    align_latex .. "\n"))
  blocks:extend(flatten_figures(right.content))
  blocks:insert(pandoc.RawBlock("latex", "\\end{minipage}\\par"))
  return blocks
end

local function format_html(element)
  local left, right = collect_panes(element)
  if not left or not right then
    return nil
  end

  local left_fraction, right_fraction = resolve_widths(element.attr)
  local gap = resolve_gap(element.attr)
  local _, valign_html = resolve_valign(element.attr)
  local _, align_html = resolve_align(element.attr)

  left = pandoc.Div(flatten_figures(left.content), left.attr)
  right = pandoc.Div(flatten_figures(right.content), right.attr)

  left.attr.attributes["style"] =
    "flex: 0 0 " .. css_width_expr(left_fraction, gap) .. "; max-width: " .. css_width_expr(left_fraction, gap) .. "; text-align: " .. align_html .. ";"
  right.attr.attributes["style"] =
    "flex: 0 0 " .. css_width_expr(right_fraction, gap) .. "; max-width: " .. css_width_expr(right_fraction, gap) .. "; text-align: " .. align_html .. ";"

  element.content = pandoc.List:new({left, right})
  element.attr.attributes["style"] =
    "display: flex; align-items: " .. valign_html .. "; gap: " .. gap .. ";"
  return element
end

return {
  {
    Div = function(element)
      if not has_class(element, "side-by-side") and not has_class(element, "sidebyside") then
        return nil
      end

      if FORMAT:match("latex") then
        return format_latex(element)
      end
      if FORMAT:match("html") then
        return format_html(element)
      end
      return nil
    end,
  },
}
