/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package spinner

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	spinnerGeneric = "generic spinner msg"
)

var _ = Describe("Spinner", func() {
	var teaModel tea.Model

	BeforeEach(func() {
		teaModel = initialModel(spinnerGeneric)
	})

	Describe("InitialModel", func() {
		It("should return a model with initialized spinner", func() {
			m, _ := teaModel.(model)
			Expect(m.spinner).NotTo(BeNil())
			Expect(m.quitting).To(BeFalse())
			Expect(m.err).To(BeNil())
		})
	})

	Describe("Init", func() {
		It("should return the spinner's Tick command", func() {
			cmd := teaModel.Init()
			Expect(cmd).NotTo(BeNil())
		})
	})

	Describe("Update", func() {
		Context("when receiving a KeyMsg", func() {
			It("should set quitting to true and return tea.Quit for Ctrl+C", func() {
				newModel, cmd := teaModel.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
				m, _ := newModel.(model)
				Expect(m.quitting).To(BeTrue())
				Expect(cmd).NotTo(BeNil())
			})

			It("should set quitting to true and return tea.Quit for 'q'", func() {
				newModel, cmd := teaModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
				m, _ := newModel.(model)
				Expect(m.quitting).To(BeTrue())
				Expect(cmd).NotTo(BeNil())
			})

			It("should set quitting to true and return tea.Quit for Esc", func() {
				newModel, cmd := teaModel.Update(tea.KeyMsg{Type: tea.KeyEsc})
				m, _ := newModel.(model)
				Expect(m.quitting).To(BeTrue())
				Expect(cmd).NotTo(BeNil())
			})

			It("should return the same model for other keys", func() {
				newModel, cmd := teaModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
				Expect(newModel).To(Equal(teaModel))
				Expect(cmd).To(BeNil())
			})
		})

		Context("when receiving an errMsg", func() {
			It("should set the error in the model", func() {
				err := errMsg(fmt.Errorf("test error"))
				newModel, cmd := teaModel.Update(err)
				m, _ := newModel.(model)
				Expect(m.err).To(MatchError("test error"))
				Expect(cmd).To(BeNil())
			})
		})

		Context("when receiving other messages", func() {
			It("should update the spinner", func() {
				newModel, cmd := teaModel.Update(spinner.TickMsg{})
				Expect(newModel).NotTo(Equal(teaModel))
				Expect(cmd).NotTo(BeNil())
			})
		})
	})

	Describe("View", func() {
		Context("when there's an error", func() {
			It("should return the error message", func() {
				m, _ := teaModel.(model)
				m.err = fmt.Errorf("test error")
				Expect(m.View()).To(Equal("test error"))
			})
		})

		Context("when quitting", func() {
			It("should return the spinner view with message", func() {
				m, _ := teaModel.(model)
				m.quitting = true
				view := m.View()
				Expect(view).To(ContainSubstring(spinnerGeneric))
				Expect(view).To(ContainSubstring(m.spinner.View()))
			})
		})

		Context("when running normally", func() {
			It("should return the spinner view with message", func() {
				m, _ := teaModel.(model)
				view := m.View()
				Expect(view).To(ContainSubstring(spinnerGeneric))
				Expect(view).To(ContainSubstring(m.spinner.View()))
			})
		})
	})
})
