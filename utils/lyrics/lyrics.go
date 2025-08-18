package lyrics

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/beevik/etree"
)

type SongLyrics struct {
	Data []struct {
		Id         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			Ttml       string `json:"ttml"`
			TtmlLocalizations       string `json:"ttmlLocalizations"`
			PlayParams struct {
				Id          string `json:"id"`
				Kind        string `json:"kind"`
				CatalogId   string `json:"catalogId"`
				DisplayType int    `json:"displayType"`
			} `json:"playParams"`
		} `json:"attributes"`
	} `json:"data"`
}

func Get(storefront, songId, lrcType, language, lrcFormat, token, mediaUserToken string) (string, error) {
	if len(mediaUserToken) < 50 {
		return "", errors.New("MediaUserToken not set")
	}

	ttml, err := getSongLyrics(songId, storefront, token, mediaUserToken, lrcType, language)
	if err != nil {
		return "", err
	}

	if lrcFormat == "ttml" {
		return ttml, nil
	}

	lrc, err := TtmlToLrc(ttml)
	if err != nil {
		return "", err
	}

	return lrc, nil
}

func getSongLyrics(songId string, storefront string, token string, userToken string, lrcType string, language string) (string, error) {
	req, err := http.NewRequest("GET",
		fmt.Sprintf("https://amp-api.music.apple.com/v1/catalog/%s/songs/%s/%s?l=%s&extend=ttmlLocalizations", storefront, songId, lrcType, language), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Origin", "https://music.apple.com")
	req.Header.Set("Referer", "https://music.apple.com/")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	cookie := http.Cookie{Name: "media-user-token", Value: userToken}
	req.AddCookie(&cookie)
	do, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer do.Body.Close()
	obj := new(SongLyrics)
	_ = json.NewDecoder(do.Body).Decode(&obj)
	if obj.Data != nil {
		if len(obj.Data[0].Attributes.Ttml) > 0 {
			return obj.Data[0].Attributes.Ttml, nil
		}
		return obj.Data[0].Attributes.TtmlLocalizations, nil
	} else {
		return "", errors.New("failed to get lyrics")
	}
}

// Use for detect if lyrics have CJK, will be replaced by transliteration if exist.
func containsCJK(s string) bool {
	for _, r := range s {
		if (r >= 0x1100 && r <= 0x11FF)    || // Hangul Jamo
			(r >= 0x2E80 && r <= 0x2EFF)   || // CJK Radicals Supplement
			(r >= 0x2F00 && r <= 0x2FDF)   || // Kangxi Radicals
			(r >= 0x2FF0 && r <= 0x2FFF)   || // Ideographic Description Characters
			(r >= 0x3000 && r <= 0x303F)   || // CJK Symbols and Punctuation
			(r >= 0x3040 && r <= 0x309F)   || // Hiragana
			(r >= 0x30A0 && r <= 0x30FF)   || // Katakana
			(r >= 0x3130 && r <= 0x318F)   || // Hangul Compatibility Jamo
			(r >= 0x31C0 && r <= 0x31EF)   || // CJK Strokes
			(r >= 0x31F0 && r <= 0x31FF)   || // Katakana Phonetic Extensions
			(r >= 0x3200 && r <= 0x32FF)   || // Enclosed CJK Letters and Months
			(r >= 0x3300 && r <= 0x33FF)   || // CJK Compatibility
			(r >= 0x3400 && r <= 0x4DBF)   || // CJK Unified Ideographs Extension A
			(r >= 0x4E00 && r <= 0x9FFF)   || // CJK Unified Ideographs
			(r >= 0xA960 && r <= 0xA97F)   || // Hangul Jamo Extended-A
			(r >= 0xAC00 && r <= 0xD7AF)   || // Hangul Syllables
			(r >= 0xD7B0 && r <= 0xD7FF)   || // Hangul Jamo Extended-B
			(r >= 0xF900 && r <= 0xFAFF)   || // CJK Compatibility Ideographs
			(r >= 0xFE30 && r <= 0xFE4F)   || // CJK Compatibility Forms
			(r >= 0xFF65 && r <= 0xFF9F)   || // Halfwidth Katakana
			(r >= 0xFFA0 && r <= 0xFFDC)   || // Halfwidth Jamo
			(r >= 0x1AFF0 && r <= 0x1AFFF) || // Kana Extended-B
			(r >= 0x1B000 && r <= 0x1B0FF) || // Kana Supplement
			(r >= 0x1B100 && r <= 0x1B12F) || // Kana Extended-A
			(r >= 0x1B130 && r <= 0x1B16F) || // Small Kana Extension
			(r >= 0x1F200 && r <= 0x1F2FF) || // Enclosed Ideographic Supplement
			(r >= 0x20000 && r <= 0x2A6DF) || // CJK Unified Ideographs Extension B
			(r >= 0x2A700 && r <= 0x2B73F) || // CJK Unified Ideographs Extension C
			(r >= 0x2B740 && r <= 0x2B81F) || // CJK Unified Ideographs Extension D
			(r >= 0x2B820 && r <= 0x2CEAF) || // CJK Unified Ideographs Extension E
			(r >= 0x2CEB0 && r <= 0x2EBEF) || // CJK Unified Ideographs Extension F
			(r >= 0x2EBF0 && r <= 0x2EE5F) || // CJK Unified Ideographs Extension I
			(r >= 0x2F800 && r <= 0x2FA1F) || // CJK Compatibility Ideographs Supplement
			(r >= 0x30000 && r <= 0x3134F) || // CJK Unified Ideographs Extension G
			(r >= 0x31350 && r <= 0x323AF) {  // CJK Unified Ideographs Extension H
			return true
		}
	}
	return false
}

func TtmlToLrc(ttml string) (string, error) {
	parsedTTML := etree.NewDocument()
	err := parsedTTML.ReadFromString(ttml)
	if err != nil {
		return "", err
	}

	// 检测是否为繁体且有简体翻译
	root := parsedTTML.FindElement("tt")
	isZhHant := false
	var zhHansTranslation *etree.Element
	if root != nil {
		if langAttr := root.SelectAttr("xml:lang"); langAttr != nil && langAttr.Value == "zh-Hant" {
			isZhHant = true
			if head := root.FindElement("head"); head != nil {
				if meta := head.FindElement("metadata"); meta != nil {
					if itunes := meta.FindElement("iTunesMetadata"); itunes != nil {
						if translations := itunes.FindElement("translations"); translations != nil {
							for _, t := range translations.FindElements("translation") {
								if t.SelectAttrValue("xml:lang", "") == "zh-Hans" {
									zhHansTranslation = t
									break
								}
							}
						}
					}
				}
			}
		}
	}
	onlyOutputSimplified := isZhHant && zhHansTranslation != nil

	var lrcLines []string
	timingAttr := parsedTTML.FindElement("tt").SelectAttr("itunes:timing")
	if timingAttr != nil {
		if timingAttr.Value == "Word" {
			lrc, err := conventSyllableTTMLToLRC(ttml)
			return lrc, err
		}
		if timingAttr.Value == "None" {
			for _, p := range parsedTTML.FindElements("//p") {
				line := p.Text()
				line = strings.TrimSpace(line)
				if line != "" {
					lrcLines = append(lrcLines, line)
				}
			}
			return strings.Join(lrcLines, "\n"), nil
		}
	}

	for _, item := range parsedTTML.FindElement("tt").FindElement("body").ChildElements() {
		for _, lyric := range item.ChildElements() {
			var h, m, s, ms int
			beginAttr := lyric.SelectAttr("begin")
			if beginAttr == nil {
				return "", errors.New("no synchronised lyrics")
			}
			beginValue := beginAttr.Value
			if strings.Contains(beginValue, ":") {
				_, err = fmt.Sscanf(beginValue, "%d:%d:%d.%d", &h, &m, &s, &ms)
				if err != nil {
					_, err = fmt.Sscanf(beginValue, "%d:%d.%d", &m, &s, &ms)
					if err != nil {
						_, err = fmt.Sscanf(beginValue, "%d:%d", &m, &s)
					}
					h = 0
				}
			} else {
				_, err = fmt.Sscanf(beginValue, "%d.%d", &s, &ms)
				h, m = 0, 0
			}
			if err != nil {
				return "", err
			}
			m += h * 60
			ms = ms / 10
			var text, transText, translitText string
			
			// 获取翻译和音译
			if len(parsedTTML.FindElement("tt").FindElements("head")) > 0 {
				if len(parsedTTML.FindElement("tt").FindElement("head").FindElements("metadata")) > 0 {
					Metadata := parsedTTML.FindElement("tt").FindElement("head").FindElement("metadata")
					if len(Metadata.FindElements("iTunesMetadata")) > 0 {
						iTunesMetadata := Metadata.FindElement("iTunesMetadata")
						
						// 优先获取简体翻译（如果存在）
						if onlyOutputSimplified {
							xpath := fmt.Sprintf("//text[@for='%s']", lyric.SelectAttr("itunes:key").Value)
							trans := zhHansTranslation.FindElement(xpath)
							if trans != nil {
								if trans.SelectAttr("text") != nil {
									transText = trans.SelectAttr("text").Value
								} else {
									var transTmp []string
									for _, span := range trans.Child {
										if c, ok := span.(*etree.CharData); ok {
											transTmp = append(transTmp, c.Data)
										} else if e, ok := span.(*etree.Element); ok {
											transTmp = append(transTmp, e.Text())
										}
									}
									transText = strings.Join(transTmp, "")
								}
							}
						} else {
							// 非繁体情况，获取第一个翻译
							if len(iTunesMetadata.FindElements("translations")) > 0 {
								if len(iTunesMetadata.FindElement("translations").FindElements("translation")) > 0 {
									xpath := fmt.Sprintf("//text[@for='%s']", lyric.SelectAttr("itunes:key").Value)
									trans := iTunesMetadata.FindElement("translations").FindElement("translation").FindElement(xpath)
									if trans != nil {
										if trans.SelectAttr("text") != nil {
											transText = trans.SelectAttr("text").Value
										} else {
											var transTmp []string
											for _, span := range trans.Child {
												if c, ok := span.(*etree.CharData); ok {
													transTmp = append(transTmp, c.Data)
												} else if e, ok := span.(*etree.Element); ok {
													transTmp = append(transTmp, e.Text())
												}
											}
											transText = strings.Join(transTmp, "")
										}
									}
								}
							}
						}
						
						// 获取音译
						if len(iTunesMetadata.FindElements("transliterations")) > 0 {
							if len(iTunesMetadata.FindElement("transliterations").FindElements("transliteration")) > 0 {
								xpath := fmt.Sprintf("text[@for='%s']", lyric.SelectAttr("itunes:key").Value)
								translit := iTunesMetadata.FindElement("transliterations").FindElement("transliteration").FindElement(xpath)
								if translit != nil {
									if translit.SelectAttr("text") != nil {
										translitText = translit.SelectAttr("text").Value
									} else {
										var translitTmp []string
										for _, span := range translit.Child {
											if c, ok := span.(*etree.CharData); ok {
												translitTmp = append(translitTmp, c.Data)
											} else if e, ok := span.(*etree.Element); ok {
												translitTmp = append(translitTmp, e.Text())
											}
										}
										translitText = strings.Join(translitTmp, "")
									}
								}
							}
						}
					}
				}
			}
			
			// 获取原文
			if lyric.SelectAttr("text") == nil {
				var textTmp []string
				for _, span := range lyric.Child {
					if _, ok := span.(*etree.CharData); ok {
						textTmp = append(textTmp, span.(*etree.CharData).Data)
					} else {
						textTmp = append(textTmp, span.(*etree.Element).Text())
					}
				}
				text = strings.Join(textTmp, "")
			} else {
				text = lyric.SelectAttr("text").Value
			}
			
			// 输出逻辑
			if onlyOutputSimplified {
				// 只输出简体翻译
				if transText != "" {
					lrcLines = append(lrcLines, fmt.Sprintf("[%02d:%02d.%02d]%s", m, s, ms, transText))
				}
			} else {
				// 输出原文
				lrcLines = append(lrcLines, fmt.Sprintf("[%02d:%02d.%02d]%s", m, s, ms, text))
				
				// 输出翻译
				if transText != "" {
					lrcLines = append(lrcLines, fmt.Sprintf("[%02d:%02d.%02d]%s", m, s, ms, transText))
				}
				
				// 输出音译（如果需要）
				if translitText != "" && containsCJK(text) {
					lrcLines = append(lrcLines, fmt.Sprintf("[%02d:%02d.%02d]%s", m, s, ms, translitText))
				}
			}
		}
	}
	return strings.Join(lrcLines, "\n"), nil
}

func conventSyllableTTMLToLRC(ttml string) (string, error) {
	parsedTTML := etree.NewDocument()
	if err := parsedTTML.ReadFromString(ttml); err != nil {
		return "", err
	}

	root := parsedTTML.FindElement("tt")
	if root == nil {
		return "", errors.New("invalid ttml")
	}

	// ------- 是否为繁体并且有简体翻译 -------
	isZhHant := false
	if a := root.SelectAttr("xml:lang"); a != nil && a.Value == "zh-Hant" {
		isZhHant = true
	}

	var zhHansTrans *etree.Element
	if isZhHant {
		if head := root.FindElement("head"); head != nil {
			if meta := head.FindElement("metadata"); meta != nil {
				if itunes := meta.FindElement("iTunesMetadata"); itunes != nil {
					if transList := itunes.FindElement("translations"); transList != nil {
						for _, t := range transList.FindElements("translation") {
							if t.SelectAttrValue("xml:lang", "") == "zh-Hans" {
								zhHansTrans = t
								break
							}
						}
					}
				}
			}
		}
	}

	var lrcLines []string
	parseTime := func(timeValue string, _ int) (string, error) {
		var h, m, s, ms int
		var err error
		if strings.Contains(timeValue, ":") {
			_, err = fmt.Sscanf(timeValue, "%d:%d:%d.%d", &h, &m, &s, &ms)
			if err != nil {
				_, err = fmt.Sscanf(timeValue, "%d:%d.%d", &m, &s, &ms)
				h = 0
			}
		} else {
			_, err = fmt.Sscanf(timeValue, "%d.%d", &s, &ms)
			h, m = 0, 0
		}
		if err != nil {
			return "", err
		}
		m += h * 60
		ms = ms / 10
		return fmt.Sprintf("[%02d:%02d.%02d]", m, s, ms), nil
	}

	divs := root.FindElement("body").FindElements("div")
	for _, div := range divs {
		for _, item := range div.ChildElements() { // 每行歌词<p>
			var (
				lrcSyllables       []string
				i                  int
				endTime            string
				translitLine       string
				transLine          string
				buildZhHansPerWord bool
			)

			// ------- 遍历逐字（原文）以便拿到 endTime，并保持原逻辑 -------
			for _, node := range item.Child {
				// 空白字符（span 之间的空格）
				if c, ok := node.(*etree.CharData); ok {
					if i > 0 {
						lrcSyllables = append(lrcSyllables, c.Data)
					}
					continue
				}
				lyric, ok := node.(*etree.Element)
				if !ok || lyric.SelectAttr("begin") == nil {
					continue
				}

				beginTime, err := parseTime(lyric.SelectAttrValue("begin", ""), i)
				if err != nil {
					return "", err
				}
				endTime, err = parseTime(lyric.SelectAttrValue("end", ""), 1)
				if err != nil {
					return "", err
				}

				var text string
				if lyric.SelectAttr("text") == nil {
					var textTmp []string
					for _, span := range lyric.Child {
						if cc, ok := span.(*etree.CharData); ok {
							textTmp = append(textTmp, cc.Data)
						} else if ee, ok := span.(*etree.Element); ok {
							textTmp = append(textTmp, ee.Text())
						}
					}
					text = strings.Join(textTmp, "")
				} else {
					text = lyric.SelectAttrValue("text", "")
				}
				lrcSyllables = append(lrcSyllables, fmt.Sprintf("%s%s", beginTime, text))

				// 第一个字时，尝试构造“翻译行”
				if i == 0 {
					// ---------- 音译行：保持原逻辑 ----------
					if head := root.FindElement("head"); head != nil {
						if meta := head.FindElement("metadata"); meta != nil {
							if itunes := meta.FindElement("iTunesMetadata"); itunes != nil {
								// 音译逐字
								if tlers := itunes.FindElement("transliterations"); tlers != nil {
									if tler := tlers.FindElement("transliteration"); tler != nil {
										xpath := fmt.Sprintf("text[@for='%s']", item.SelectAttrValue("itunes:key", ""))
										if trans := tler.FindElement(xpath); trans != nil {
											var parts []string
											var start string
											for _, ch := range trans.Child {
												if e, ok := ch.(*etree.Element); ok && e.Tag == "span" {
													beg := e.SelectAttrValue("begin", "")
													if beg == "" {
														continue
													}
													ts, err := parseTime(beg, 2)
													if err != nil {
														return "", err
													}
													if start == "" {
														start, _ = parseTime(beg, -1)
													}
													parts = append(parts, fmt.Sprintf("%s%s", ts, e.Text()))
												} else if cd, ok := ch.(*etree.CharData); ok {
													// 保留中间的空格
													parts = append(parts, cd.Data)
												}
											}
											if len(parts) > 0 {
												translitLine = start + strings.Join(parts, "")
											}
										}
									}
								}

								// ---------- 翻译行：若繁体且有简体翻译 → 逐字拼接 ----------
								var transContainer *etree.Element
								if zhHansTrans != nil {
									transContainer = zhHansTrans
									buildZhHansPerWord = true
								} else if trs := itunes.FindElement("translations"); trs != nil {
									// 非繁体场景：保持原逻辑，取第一个 translation
									transContainer = trs.FindElement("translation")
								}

								if transContainer != nil {
									xpath := fmt.Sprintf("//text[@for='%s']", item.SelectAttrValue("itunes:key", ""))
									if tx := transContainer.FindElement(xpath); tx != nil {
										if buildZhHansPerWord {
											// 按 <span begin> 逐字拼接
											var parts []string
											var start string
											hasSpan := false
											for _, ch := range tx.Child {
												switch v := ch.(type) {
												case *etree.Element:
													if v.Tag == "span" {
														beg := v.SelectAttrValue("begin", "")
														if beg == "" {
															continue
														}
														ts, err := parseTime(beg, 2)
														if err != nil {
															return "", err
														}
														if start == "" {
															start, _ = parseTime(beg, -1)
														}
														parts = append(parts, fmt.Sprintf("%s%s", ts, v.Text()))
														hasSpan = true
													} else {
														parts = append(parts, v.Text())
													}
												case *etree.CharData:
													// 保留空格等
													if v.Data != "" {
														parts = append(parts, v.Data)
													}
												}
											}
											if hasSpan {
												transLine = strings.Join(parts, "")
											} else {
												// 极端兜底：没有 span 就退回整句
												txt := tx.SelectAttrValue("text", "")
												if txt == "" {
													var tmp []string
													for _, ch := range tx.Child {
														if cd, ok := ch.(*etree.CharData); ok {
															tmp = append(tmp, cd.Data)
														} else if ee, ok := ch.(*etree.Element); ok {
															tmp = append(tmp, ee.Text())
														}
													}
													txt = strings.Join(tmp, "")
												}
												start, _ := parseTime(lyric.SelectAttrValue("begin", ""), -1)
												transLine = start + txt
											}
										} else {
											// 原逻辑：非繁体时保持整句
											var txt string
											if tx.SelectAttr("text") == nil {
												var tmp []string
												for _, ch := range tx.Child {
													if cd, ok := ch.(*etree.CharData); ok {
														tmp = append(tmp, cd.Data)
													} else if ee, ok := ch.(*etree.Element); ok {
														tmp = append(tmp, ee.Text())
													}
												}
												txt = strings.Join(tmp, "")
											} else {
												txt = tx.SelectAttrValue("text", "")
											}
											start, _ := parseTime(lyric.SelectAttrValue("begin", ""), -1)
											transLine = start + txt
										}
									}
								}
							}
						}
					}
				}

				i++
			}

			// ------- 输出阶段 -------
			if isZhHant && zhHansTrans != nil {
				// 只保留简体逐字翻译行
				if transLine != "" {
					lrcLines = append(lrcLines, transLine+endTime)
				}
				// 繁体原文/音译不输出
				continue
			}

			// 非繁体：保持原有算法与顺序
			if transLine != "" {
				lrcLines = append(lrcLines, strings.Join(lrcSyllables, "")+endTime)
				lrcLines = append(lrcLines, transLine+endTime)
			} else {
				lrcLines = append(lrcLines, strings.Join(lrcSyllables, "")+endTime)
			}
			if translitLine != "" && containsCJK(strings.Join(lrcSyllables, "")) {
				lrcLines = append(lrcLines, translitLine+endTime)
			}
		}
	}

	return strings.Join(lrcLines, "\n"), nil
}
